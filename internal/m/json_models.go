package m

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strconv"
)

type US_CongressLegislator struct {
	Id   CongIdentifiers `json:"id"`
	Name struct {
		First    string `json:"first"`
		Last     string `json:"last"`
		Official string `json:"official_full"`
	} `json:"name"`
	Bio struct {
		Birthday string `json:"birthday"`
		Gender   string `json:"gender"`
	} `json:"bio"`
	Terms []Terms `json:"terms"`
}

func (l US_CongressLegislator) ServedDuringYear(year int) bool {
	for _, term := range l.Terms {
		start, err := parseYear(term.Start)
		if err != nil {
			continue
		}
		end, err := parseYear(term.End)
		if err != nil {
			continue
		}
		if start <= year && year <= end {
			return true
		}
	}
	return false
}

func (l US_CongressLegislator) ServedAsDuringYear(role string, year int) bool {
	for _, term := range l.Terms {
		start, err := parseYear(term.Start)
		if err != nil {
			continue
		}
		end, err := parseYear(term.End)
		if err != nil {
			continue
		}
		if start <= year && year <= end && term.Type == role {
			return true
		}
	}
	return false
}

/**
 * Grabs the year from a date string formatted as YYYY-MM-DD
 */
func parseYear(date string) (int, error) {
	return strconv.Atoi(date[:4])
}

// Implement the scanner interface for US_CongressLegislators
func (l *US_CongressLegislator) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan US_CongressLegislators")
	}
	return json.Unmarshal(b, l)
}

// Implement the value interface for US_CongressLegislators
func (l US_CongressLegislator) Value() (driver.Value, error) {
	return json.Marshal(l)
}

// See https://github.com/unitedstates/congress-legislators?tab=readme-ov-file
// For the data dictionary of each of these fields
type CongIdentifiers struct {
	Bioguide       string   `json:"bioguide"`
	Fec            []string `json:"fec"`
	Cspan          int      `json:"cspan"`
	Wikipedia      string   `json:"wikipedia"`
	HouseHistory   int      `json:"house_history"`
	Ballotpedia    string   `json:"ballotpedia"`
	Maplight       int      `json:"maplight"`
	Icpsr          int      `json:"icpsr"`
	Wikidata       string   `json:"wikidata"`
	GoogleEntityID string   `json:"google_entity_id"`
}

type Terms struct {
	Type     string `json:"type"`
	State    string `json:"state"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Party    string `json:"party"`
	District int    `json:"district"`
}

func (t Terms) IsSenator() bool {
	return t.Type == "sen"
}
