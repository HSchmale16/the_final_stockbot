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

	RollCallNum      int       `gorm:"uniqueIndex:,unique,composite:votename"`
	CongressNum      int       `gorm:"uniqueIndex:,unique,composite:votename"`
	Session          string    `gorm:"uniqueIndex:,unique,composite:votename"`
	Chamber          string    `gorm:"uniqueIndex:,unique,composite:votename"`
	ActionAt         time.Time `gorm:"index"`
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

	VoteID     uint   `gorm:"index;index:idx_vote_records_vote_member_status,priority:1;uniqueIndex:unique,composite:voterecord"`
	MemberId   string `gorm:"index;index:idx_vote_records_vote_member_status,priority:2;uniqueIndex:unique,composite:voterecord"`
	VoteStatus string `gorm:"index:idx_vote_records_vote_member_status,priority:3"`

	Vote   Vote
	Member m.DB_CongressMember
}
