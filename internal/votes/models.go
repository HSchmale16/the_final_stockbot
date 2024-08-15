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

	RollCallNum      int
	CongressNum      int
	Chamber          string
	ActionAt         time.Time
	VoteType         string
	LegisName        string
	VoteResult       string
	AmmendmentNum    int
	AmmendmentAuthor string

	VoteRecords []VoteRecord
}

type VoteRecord struct {
	ID int `gorm:"primaryKey"`

	VoteID     int    `gorm:"index"`
	MemberId   string `gorm:"index"`
	VoteStatus string

	Vote   Vote
	Member m.DB_CongressMember
}
