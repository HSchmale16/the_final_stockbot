package stocks

import (
	"database/sql"
	"time"

	"github.com/hschmale16/the_final_stockbot/internal/m"
)

func init() {
	m.RegisterModels(&FinDisclosureDocument{})
	m.RegisterModels(&FinTransaction{})
}

type DB_CongressMember = m.DB_CongressMember

type FinDisclosureDocument struct {
	ID        uint
	CreatedAt time.Time
	UpdatedAt time.Time

	MemberId sql.NullString
	Filing

	Processed string // status of processing

	Member DB_CongressMember
}

func (f *FinDisclosureDocument) TableName() string {
	return "fin_disclosure_documents"
}

type FinTransaction struct {
	ID        uint
	CreatedAt time.Time
	UpdatedAt time.Time

	Date               time.Time
	Stock              string
	Company            string
	Description        string
	AmountCategory     string
	FilingStatus       string
	CapGainsGreater200 bool

	FinDisclosureDocumentID uint
	FinDisclosureDocument   FinDisclosureDocument
}
