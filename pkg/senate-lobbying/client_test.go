package senatelobbying_test

import (
	"testing"

	senatelobbying "github.com/hschmale16/the_final_stockbot/pkg/senate-lobbying"
)

func TestMain(t *testing.T) {

}

func TestGetContributionListUrl(t *testing.T) {
	url := senatelobbying.GetContributionListUrl(senatelobbying.ContributionListingFilterParams{
		FilingYear: "2021",
	})

	if url != "https://lda.senate.gov/api/v1/contributions/?filing_year=2021" {
		t.Errorf("Expected url to be https://lda.senate.gov/api/v1/contributions/?filing_year=2021, got %s", url)
	}
}
