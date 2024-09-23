package votes

import (
	"embed"
	"time"

	"github.com/hschmale16/the_final_stockbot/internal/m"
	"gorm.io/gorm"
)

//go:embed html_templates/*
var templateFS embed.FS

func init() {
	m.RegisterModels(&Vote{}, &VoteRecord{})
	m.RegisterDebugFilePath("internal/votes/html_templates")
	m.RegisterEmbededFS(templateFS)
}

type Vote struct {
	gorm.Model

	RollCallNum      int    `gorm:"uniqueIndex:,unique,composite:votename"`
	CongressNum      int    `gorm:"uniqueIndex:,unique,composite:votename"`
	Session          string `gorm:"uniqueIndex:,unique,composite:votename"`
	Chamber          string `gorm:"uniqueIndex:,unique,composite:votename"`
	ActionAt         time.Time
	VoteType         string
	LegisName        string
	VoteResult       string
	AmmendmentNum    int
	AmmendmentAuthor string
	VoteDesc         string
	Url              string `gorm:uniqueIndex`

	VoteRecords []VoteRecord
}

type VoteRecord struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time

	VoteID     uint   `gorm:"uniqueIndex:unique,composite:voterecord"`
	MemberId   string `gorm:"uniqueIndex:unique,composite:voterecord"`
	VoteStatus string

	Vote   Vote
	Member m.DB_CongressMember
}
