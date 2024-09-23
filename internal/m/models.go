package m

import (
	"fmt"
	"log"
	"os"
	"time"

	fecwrangling "github.com/hschmale16/the_final_stockbot/internal/fecwrangling"
	"github.com/hschmale16/the_final_stockbot/internal/lobbying"
	"github.com/ncruces/go-sqlite3"
	"github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/ncruces/go-sqlite3/gormlite"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/prometheus"
)

type GovtRssItem struct {
	gorm.Model
	DescriptiveMetaUrl string
	FullTextUrl        string
	Title              string
	Link               string    `gorm:"index:,unique,composite:unique_per_item"`
	PubDate            time.Time `gorm:"index:,unique,composite:unique_per_item"`
	ProcessedOn        time.Time
	Metadata           LawModsData `gorm:"type:jsonb"`

	// many to many relationship of tags through GovtRssItemTag
	Tags       []Tag                  `gorm:"many2many:govt_rss_item_tag;"`
	Categories []Tag                  `gorm:"many2many:rss_category"`
	Sponsors   []DB_CongressMember    `gorm:"many2many:congress_member_sponsored;"`
	Committees []DB_CongressCommittee `gorm:"many2many:db_congress_committee_govt_rss_items;"`
}

func (GovtRssItem) TableName() string {
	return "govt_rss_item"
}

type stupidPair struct{ Num, Percent float64 }
type SponsorshipMap map[string]stupidPair

func (g GovtRssItem) ComputeSponsorship() SponsorshipMap {
	return MakeSponsorshipMap(g.Sponsors, func(s DB_CongressMember) string {
		return s.Party()
	})
}

func MakeSponsorshipMap[K any](items []K, toString func(K) string) SponsorshipMap {
	sponsorship := make(SponsorshipMap)

	for _, item := range items {
		str := toString(item)
		entry, ok := sponsorship[str]
		if ok {
			entry.Num++
		} else {
			entry.Num = 1
		}
		sponsorship[str] = entry
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

	// Colorize the tags returned
	CssColor string `gorm:"default:'bg-secondary'"`

	// If true do not show by default in the tag list per bill
	Hidden bool
}

func (Tag) TableName() string {
	return "tag"
}

type TagUse struct {
	ID        uint
	CreatedAt time.Time
	TagId     uint
	IpAddr    string
	UserAgent string
	UseType   string

	Tag Tag `gorm:"foreignKey:TagId"`
}

func (TagUse) TableName() string {
	return "tag_use"
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

/**
 * SearchQuery is a record of a search query done on the front page
 */
type SearchQuery struct {
	ID         uint
	CreatedAt  time.Time
	IpAddr     string
	UserAgent  string
	Query      string
	NumResults int
	FtsResults int
}

func (SearchQuery) TableName() string {
	return "search_query"
}

///////////////////////////////////////////////////////////////////

type DB_CongressMember struct {
	BioGuideId         string `gorm:"primaryKey"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
	CongressMemberInfo US_CongressLegislator `gorm:"type:jsonb"`
	Name               string
	Sponsored          []GovtRssItem            `gorm:"many2many:congress_member_sponsored;"`
	Committees         []DB_CommitteeMembership `gorm:"foreignKey:CongressMemberId"`
}

func (d DB_CongressMember) TableName() string {
	return "congress_member"
}

func (d DB_CongressMember) TookOfficeOn() string {
	return d.CongressMemberInfo.Terms[0].Start
}

func (d DB_CongressMember) Party() string {
	return d.CongressMemberInfo.Terms[len(d.CongressMemberInfo.Terms)-1].Party
}

func (d DB_CongressMember) State() string {
	return d.CongressMemberInfo.Terms[len(d.CongressMemberInfo.Terms)-1].State
}

func (d DB_CongressMember) District() int {
	return d.CongressMemberInfo.Terms[len(d.CongressMemberInfo.Terms)-1].District
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
	ID                          uint
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

type FeedbackItem struct {
	ID        uint
	CreatedAt time.Time
	UpdatedAt time.Time

	Name      string
	Email     string
	Url       string
	Message   string
	UserAgent string
	IpAddr    string

	Status string `gorm:"default:'unanswered'"`
}

func (f FeedbackItem) TableName() string {
	return "feedback_items"
}

///////////////////////////////////////////////////////////////////

func GetLogger() logger.Interface {
	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             70 * time.Millisecond, // Slow SQL threshold
			LogLevel:                  logger.Silent,         // Log level
			IgnoreRecordNotFoundError: true,                  // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      false,                 // Don't include params in the SQL log
			Colorful:                  true,                  // Disable color
		},
	)
}

func GetSqliteDB() (*gorm.DB, error) {
	conn, err := driver.Open("congress.sqlite", func(conn *sqlite3.Conn) error {
		return nil
	})
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(gormlite.OpenDB(conn), &gorm.Config{
		Logger:      GetLogger(),
		PrepareStmt: true,
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func GetPostgresqlDB() (*gorm.DB, error) {
	whoami := os.Getenv("USER")
	dsn := fmt.Sprintf("host=localhost user=%s dbname=congress sslmode=disable", whoami)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: GetLogger(),
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}

/**
 * Sets up the stupid database
 */
func SetupDB() (*gorm.DB, error) {
	db, err := GetPostgresqlDB()
	if err != nil {
		panic(err)
	}

	if err := ApplyMigrations(db); err != nil {
		return nil, err
	}

	return db, nil
}

func ApplyMigrations(db *gorm.DB) error {
	db.Use(prometheus.New(prometheus.Config{
		DBName:          "congress", // use `DBName` as metrics label
		RefreshInterval: 15,         // Refresh metrics interval (default 15 seconds)
		StartServer:     true,       // start http server to expose metrics
		HTTPServerPort:  2112,       // configure http server port, default port 8080 (if you have configured multiple instances, only the first `HTTPServerPort` will be used to start server)
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.Postgres{},
		}, // user defined metrics
	}))

	// Auto migrate models
	if err := db.AutoMigrate(&GovtRssItem{}, &GovtLawText{}, &Tag{}, &GovtRssItemTag{}, &GenerationError{}, &RssCategory{}, &LawOffset{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&FederalRegisterItem{}, &FeedbackItem{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&SearchQuery{}, &DB_CongressMember{}, &CongressMemberSponsored{}, &TagUse{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(fecwrangling.CampaignCanidateLinkage{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(lobbying.LobbyingSqlQuery{}); err != nil {
		return err
	}

	// Register additional models
	for i, m := range additionalModels {
		log.Printf("Registering model %d: %T", i, m)
		if err := db.AutoMigrate(m); err != nil {
			return err
		}
	}

	// Check if some full text search tables exist
	// if !db.Migrator().HasTable("fts_law_title") {
	// 	log.Print("Creating FTS table")
	// 	if err := db.Exec("CREATE VIRTUAL TABLE fts_law_title USING fts5(title, pub_date, content='govt_rss_item', content_rowid='id');").Error; err != nil {
	// 		return nil, err
	// 	}
	// 	if err := db.Exec("CREATE TRIGGER trg_fts_law_title AFTER INSERT ON govt_rss_item BEGIN INSERT INTO fts_law_title(rowid, title, pub_date) VALUES (new.id, new.title, new.pub_date); END;").Error; err != nil {
	// 		return nil, err
	// 	}
	// }

	if err := db.AutoMigrate(&DB_CongressCommittee{}, &DB_CommitteeMembership{}); err != nil {
		return err
	}

	return nil
}

func GetTag(db *gorm.DB, tagName string) Tag {
	tag := Tag{Name: cases.Title(language.Und).String(tagName)}

	db.FirstOrCreate(&tag, tag)
	// fmt.Println("Tag:", tagName, " --> ", tag)

	return tag
}

var additionalModels = make([]interface{}, 0)

func RegisterModels(models ...interface{}) {
	additionalModels = append(additionalModels, models...)
	// for i, m := range additionalModels {
	// 	log.Printf("Registering model %d: %T", i, m)
	// }
}
