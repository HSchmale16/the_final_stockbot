package lobbying

import "time"

type LobbyingSqlQuery struct {
	ID         uint `gorm:"primaryKey"`
	CreatedAt  time.Time
	SqlText    string  `gorm:"type:text"`
	ErrorText  *string `gorm:"type:text"`
	NumResults int
	IpAddr     string
	UserAgent  string
}
