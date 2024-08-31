package congress

import (
	"time"

	"github.com/hschmale16/the_final_stockbot/internal/m"
	"gorm.io/datatypes"
)

func init() {
	m.RegisterModels(&Bill{})
}

type Bill struct {
	CongressNumber int
	BillNumber     int
	BillType       string

	CreatedAt time.Time
	UpdatedAt time.Time
	ID        uint `gorm:"primaryKey"`

	JsonBlob datatypes.JSON
}

func (b Bill) TableName() string {
	return "bills"
}
