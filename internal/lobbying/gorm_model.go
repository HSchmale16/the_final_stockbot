package lobbying

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type LobbyingSqlQuery struct {
	ID         uint `gorm:"primaryKey"`
	CreatedAt  time.Time
	SqlText    string  `gorm:"type:text"`
	ErrorText  *string `gorm:"type:text"`
	NumResults int
	IpAddr     string
	UserAgent  string
}

func LogAnalytics(sql string, err error, i int, c *fiber.Ctx) {
	db := c.Locals("db").(*gorm.DB)

	analytics := LobbyingSqlQuery{
		SqlText:    sql,
		ErrorText:  shittyString(err),
		NumResults: i,
		IpAddr:     c.IP(),
		UserAgent:  c.Get("User-Agent"),
	}

	db.Create(&analytics)
}
