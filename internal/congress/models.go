package congress

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hschmale16/the_final_stockbot/internal/m"
	"github.com/hschmale16/the_final_stockbot/internal/votes"
	"gorm.io/datatypes"
)

func init() {
	m.RegisterModels(Bill{}, BillAction{}, BillCosponsor{})
}

type Bill struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	FetchedAt         time.Time
	CosponsorsFetchAt time.Time
	ActionsFetchAt    time.Time
	CommitteeFetchAt  time.Time

	CongressNumber int    `gorm:"uniqueIndex:idx_bill_congress_number_bill_number"`
	BillNumber     string `gorm:"uniqueIndex:idx_bill_congress_number_bill_number"`
	BillType       string `gorm:"uniqueIndex:idx_bill_congress_number_bill_number"`
	Title          string
	JsonBlob       datatypes.JSON

	Cosponsors []BillCosponsor
	Actions    []BillAction
	Committees []m.DB_CongressCommittee `gorm:"many2many:bill_committees;"`
}

func (b Bill) TableName() string {
	return "bills"
}

func (b Bill) FormatTitle() string {
	return fmt.Sprintf("%d %s %s - %s", b.CongressNumber, b.BillType, b.BillNumber, b.Title)
}

func (b Bill) GetJsonBlob() map[string]interface{} {
	var blob map[string]interface{}
	err := json.Unmarshal(b.JsonBlob, &blob)
	if err != nil {
		return nil
	}
	return blob
}

//////////////////////////////////////////////////////////////////////////////////////////////

type BillAction struct {
	ID        uint
	CreatedAt time.Time
	UpdatedAt time.Time

	ActionTime       time.Time
	ActionCode       string
	ActionDate       string
	SourceSystemCode int
	Type             string
	Text             string

	VoteId *uint
	Vote   *votes.Vote

	BillID      uint
	CommitteeId sql.NullString

	Bill      Bill
	Committee m.DB_CongressCommittee
}

func (b BillAction) TableName() string {
	return "bill_actions"
}

type BillCosponsor struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	BillID            uint   `gorm:"uniqueIndex:idx_cosponsor_bill_id_member_id"`
	MemberId          string `gorm:"uniqueIndex:idx_cosponsor_bill_id_member_id"`
	OriginalCosponsor bool
	SponsorshipDate   string
	IsSponsor         bool

	Bill   Bill
	Member m.DB_CongressMember
}

func (b BillCosponsor) TableName() string {
	return "bill_cosponsors"
}
