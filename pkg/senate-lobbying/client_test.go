package senatelobbying_test

import (
	_ "embed"
	"testing"

	"encoding/json"

	senatelobbying "github.com/hschmale16/the_final_stockbot/pkg/senate-lobbying"
	"github.com/stretchr/testify/assert"
)

//go:embed test_data/filings.json
var filingsData []byte

func TestLoad_FilingData(t *testing.T) {
	var filings senatelobbying.FilingListResponse
	err := json.Unmarshal(filingsData, &filings)

	assert.Nil(t, err)
	assert.Equal(t, 1739337, filings.Count)
	assert.Equal(t, 25, len(filings.Results))
}

func TestGetContributionListUrl(t *testing.T) {
	url := senatelobbying.GetContributionListUrl(senatelobbying.ContributionListingFilterParams{
		FilingYear: "2021",
	})

	if url != "https://lda.senate.gov/api/v1/contributions/?filing_year=2021" {
		t.Errorf("Expected url to be https://lda.senate.gov/api/v1/contributions/?filing_year=2021, got %s", url)
	}
}
