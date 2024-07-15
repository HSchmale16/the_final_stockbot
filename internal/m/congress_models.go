package m

import (
	"time"
)

// https://theunitedstates.io/congress-legislators/committees-current.json
type F_CongressCommittee struct {
	Type               string `json:"type"`
	Name               string `json:"name"`
	URL                string `json:"url"`
	ThomasId           string `json:"thomas_id" gorm:"primaryKey"`
	HouseCommitteeId   string `json:"house_committee_id"`
	SenateCommitteeId  string `json:"senate_committee_id"`
	Jurisdiction       string `json:"jurisdiction"`
	JurisdictionSource string `json:"jurisdiction_source"`
	MinorityUrl        string `json:"minority_url"`
	RssUrl             string `json:"rss_url"`
	MinorityRssUrl     string `json:"minority_rss_url"`

	Wikipedia string `json:"wikipedia"`
	YoutubeId string `json:"youtube_id"`
}

type JSON_CongressCommittee struct {
	F_CongressCommittee
	Subcommittees []F_CongressCommittee `json:"subcommittees"`
}

// type Subcommittee struct {
// 	Name     string `json:"name"`
// 	ThomasId string `json:"thomas_id" gorm:"primaryKey"`
// 	Phone    string `json:"phone"`
// 	Address  string `json:"address"`
// }

type DB_CongressCommittee struct {
	CreatedAt time.Time
	UpdatedAt time.Time

	F_CongressCommittee

	Subcommittees     []DB_CongressCommittee   `gorm:"foreignKey:ParentCommitteeId"`
	ParentCommitteeId *string                  `gorm:"index"`
	Memberships       []DB_CommitteeMembership `gorm:"foreignKey:CommitteeId"`
}

func (DB_CongressCommittee) TableName() string {
	return "congress_committee"
}

type DB_CommitteeMembership struct {
	CreatedAt time.Time
	UpdatedAt time.Time

	// For some retarded reason gorm wants these columns named this way.
	CongressMemberId string `gorm:"primaryKey;column:db_congress_member_bio_guide_id"`
	CommitteeId      string `gorm:"primaryKey;column:db_congress_committee_thomas_id"`
	Rank             int
	Title            string

	CongressMember DB_CongressMember    `gorm:"foreignKey:CongressMemberId"`
	Committee      DB_CongressCommittee `gorm:"foreignKey:CommitteeId"`
}

func (DB_CommitteeMembership) TableName() string {
	return "db_committee_memberships"
}
