package lobbying

import "time"

type LobbyingSqlQuery struct {
	CreatedAt  time.Time
	UID        int64
	SqlText    string  `gorm:"type:text"`
	ErrorText  *string `gorm:"type:text"`
	NumResults int
	IpAddr     string
	UserAgent  string
}
