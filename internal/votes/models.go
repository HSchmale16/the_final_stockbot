package votes

import (
	"time"

	"github.com/hschmale16/the_final_stockbot/internal/m"
	"gorm.io/gorm"
)

func init() {
	m.RegisterModels(&Vote{}, &VoteRecord{})
}

type Vote struct {
	gorm.Model

	RollCallNum      int    `gorm:"index:,unique,composite:votename"`
	CongressNum      int    `gorm:"index:,unique,composite:votename"`
	Session          string `gorm:"index:,unique,composite:votename"`
	Chamber          string `gorm:"index:,unique,composite:votename"`
	ActionAt         time.Time
	VoteType         string
	LegisName        string
	VoteResult       string
	AmmendmentNum    int
	AmmendmentAuthor string
	VoteDesc         string

	VoteRecords []VoteRecord
}

type VoteRecord struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time

	VoteID     uint   `gorm:"index:unique,composite:voterecord"`
	MemberId   string `gorm:"index:unique,composite:voterecord"`
	VoteStatus string

	Vote   Vote
	Member m.DB_CongressMember
}
