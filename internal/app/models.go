package app

import (
	"log"
	"os"
	"time"

	fecwrangling "github.com/hschmale16/the_final_stockbot/internal/fecwrangling"
	"github.com/hschmale16/the_final_stockbot/internal/lobbying"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type GovtRssItem struct {
	gorm.Model
	DescriptiveMetaUrl string
	FullTextUrl        string
	Title              string
	Link               string    `gorm:"index:,unique,composite:unique_per_item"`
	PubDate            time.Time `gorm:"index:,unique,composite:unique_per_item"`
	ProcessedOn        time.Time
	Metadata           LawModsData

	// many to many relationship of tags through GovtRssItemTag
	Tags       []Tag               `gorm:"many2many:govt_rss_item_tag;"`
	Categories []Tag               `gorm:"many2many:rss_category"`
	Sponsors   []DB_CongressMember `gorm:"many2many:congress_member_sponsored;"`
}

func (GovtRssItem) TableName() string {
	return "govt_rss_item"
}

type stupidPair struct{ Num, Percent float64 }
type SponsorshipMap map[string]stupidPair

func (g GovtRssItem) ComputeSponsorship() SponsorshipMap {
	sponsorship := make(SponsorshipMap)

	for _, sponsor := range g.Sponsors {
		p := sponsor.Party()
		entry, ok := sponsorship[p]
		if ok {
			entry.Num++
		} else {
			entry.Num = 1
		}
		sponsorship[p] = entry
	}

	// Compute Sum
	sum := 0.0
	for _, v := range sponsorship {
		sum += v.Num
	}

	// Normalize
	for k, v := range sponsorship {
		sponsorship[k] = stupidPair{v.Num, v.Num / sum * 100}
	}

	return sponsorship
}

type FederalRegisterItem struct {
	gorm.Model
	GovtRssItemId uint
	Type          string
	FullText      string

	// many to many relationship of tags through FederalRegisterTag
	Tags []Tag `gorm:"many2many:federal_register_tag;"`
}

func (FederalRegisterItem) TableName() string {
	return "federal_register_item"
}

type FederalRegisterTag struct {
	FederalRegisterItemId uint `gorm:"index:,unique,composite:myname"`
	TagId                 uint `gorm:"index:,unique,composite:myname"`
}

func (FederalRegisterTag) TableName() string {
	return "federal_register_tag"
}

/**
 * Create a 2nd relationship to cover built in categories
 */
type RssCategory struct {
	GovtRssItemId uint `gorm:"index:,unique,composite:unique_per_item"`
	TagId         uint `gorm:"index:,unique,composite:unique_per_item"`
}

/**
 * GovtLawText is the full text of a law item fetched from the FullTextUrl
 */
type GovtLawText struct {
	gorm.Model

	GovtRssItemId uint
	GovtRssItem   GovtRssItem
	Text          string
	ModsXML       string
}

func (GovtLawText) TableName() string {
	return "govt_law_text"
}

/** Tag is a simple tag for categorizing items */
type Tag struct {
	ID        uint
	CreatedAt time.Time
	Name      string `gorm:"uniqueIndex"`

	// ShortLine is a short version of the tag name
	ShortLine string

	// If true do not show by default in the tag list per bill
	Hidden bool
}

func (Tag) TableName() string {
	return "tag"
}

/** GovtRssItemTag is a many-to-many relationship between GovtRssItem and Tag */
type GovtRssItemTag struct {
	CreatedAt  time.Time
	ModifiedAt time.Time
	ID         uint

	GovtRssItemId uint `gorm:"index:,unique,composite:myname"`
	TagId         uint `gorm:"index:,unique,composite:myname"`

	GovtRssItem GovtRssItem
	Tag         Tag
	LawOffsets  []LawOffset
}

type LawOffset struct {
	ID               uint
	GovtRssItemTagId uint
	CreatedAt        time.Time
	Offset           int
}

// Add a compound unique index on GovtRssItemId and TagId
func (GovtRssItemTag) TableName() string {
	return "govt_rss_item_tag"
}

type GenerationError struct {
	ID            uint
	CreatedAt     time.Time
	ErrorMessage  string `gorm:"type:text"`
	AttemptedText string `gorm:"type:text"`
	Model         string
	Source        string
}

func (GenerationError) TableName() string {
	return "generation_error"
}

type SearchQuery struct {
	ID         uint
	CreatedAt  time.Time
	Query      string
	NumResults int
}

func (SearchQuery) TableName() string {
	return "search_query"
}

///////////////////////////////////////////////////////////////////

type DB_CongressMember struct {
	BioGuideId         string `gorm:"primaryKey"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
	CongressMemberInfo US_CongressLegislator
	Name               string
	Sponsored          []GovtRssItem `gorm:"many2many:congress_member_sponsored;"`
}

func (d DB_CongressMember) TableName() string {
	return "congress_member"
}

func (d DB_CongressMember) Party() string {
	return d.CongressMemberInfo.Terms[len(d.CongressMemberInfo.Terms)-1].Party
}

func (d DB_CongressMember) State() string {
	return d.CongressMemberInfo.Terms[len(d.CongressMemberInfo.Terms)-1].State
}

func (d DB_CongressMember) IsActiveMember() bool {
	currentTerm := d.CongressMemberInfo.Terms[len(d.CongressMemberInfo.Terms)-1]
	now := time.Now()
	// conver to time.time
	termEnd, _ := time.Parse("2006-01-02", currentTerm.End)
	termStart, _ := time.Parse("2006-01-02", currentTerm.Start)

	return now.After(termStart) && now.Before(termEnd)
}

func (d DB_CongressMember) Role() string {
	return d.CongressMemberInfo.Terms[len(d.CongressMemberInfo.Terms)-1].Type
}

func (d DB_CongressMember) IsSenator() bool {
	return d.Role() == "sen"
}

type CongressMemberSponsored struct {
	CreatedAt                   time.Time
	CongressNumber              string
	Chamber                     string
	Role                        string
	DB_CongressMemberBioGuideId string `gorm:"index:,unique,composite:unique_per_item"`
	GovtRssItemId               uint   `gorm:"index:,unique,composite:unique_per_item"`
}

func (CongressMemberSponsored) TableName() string {
	return "congress_member_sponsored"
}

///////////////////////////////////////////////////////////////////

/**
 * Sets up the stupid database
 */
func SetupDB() (*gorm.DB, error) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             100 * time.Millisecond, // Slow SQL threshold
			LogLevel:                  logger.Silent,          // Log level
			IgnoreRecordNotFoundError: true,                   // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,                   // Don't include params in the SQL log
			Colorful:                  false,                  // Disable color
		},
	)

	// Globally mode
	db, err := gorm.Open(sqlite.Open("congress.sqlite"), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, err
	}

	// Auto migrate models
	if err := db.AutoMigrate(&GovtRssItem{}, &GovtLawText{}, &Tag{}, &GovtRssItemTag{}, &GenerationError{}, &RssCategory{}, &LawOffset{}); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&FederalRegisterItem{}); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&SearchQuery{}, &DB_CongressMember{}, &CongressMemberSponsored{}); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(fecwrangling.CampaignCanidateLinkage{}); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(lobbying.LobbyingSqlQuery{}); err != nil {
		return nil, err
	}

	return db, nil
}

func GetTag(db *gorm.DB, tagName string) Tag {
	tag := Tag{Name: cases.Title(language.Und).String(tagName)}

	db.FirstOrCreate(&tag, tag)
	// fmt.Println("Tag:", tagName, " --> ", tag)

	return tag
}
