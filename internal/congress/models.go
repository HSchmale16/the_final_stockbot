package congress

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/hschmale16/the_final_stockbot/internal/m"
	"gorm.io/datatypes"
)

func init() {
	m.RegisterModels(Bill{}, BillAction{}, BillCosponsor{})
}

type Bill struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	CongressNumber int    `gorm:"uniqueIndex:idx_bill_congress_number_bill_number"`
	BillNumber     string `gorm:"uniqueIndex:idx_bill_congress_number_bill_number"`
	BillType       string `gorm:"uniqueIndex:idx_bill_congress_number_bill_number"`
	Title          string
	JsonBlob       datatypes.JSON

	Cosponsors []BillCosponsor
	Actions    []BillAction
}

func (b Bill) TableName() string {
	return "bills"
}

func (b Bill) FormatTitle() string {
	return fmt.Sprintf("%d %s %s - %s", b.CongressNumber, b.BillType, b.BillNumber, b.Title)
}

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
