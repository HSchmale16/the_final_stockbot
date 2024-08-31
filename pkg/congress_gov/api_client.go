package congressgov

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type CongGovApiClient struct {
	BaseURL    string
	ApiKey     string
	HTTPClient *http.Client
}

func NewClient(apiKey string) *CongGovApiClient {
	return &CongGovApiClient{
		BaseURL:    "https://api.congress.gov/",
		ApiKey:     apiKey,
		HTTPClient: &http.Client{},
	}
}

func (c *CongGovApiClient) FormatUrl(path string) string {
	return c.BaseURL + path + "?api_key=" + c.ApiKey + "&format=json&limit=250"
}

func (c *CongGovApiClient) GetLatestBillActions() (LatestBillActions, error) {
	url := c.FormatUrl("v3/bill")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return LatestBillActions{}, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return LatestBillActions{}, err
	}
	defer resp.Body.Close()

	var billActions LatestBillActions
	err = json.NewDecoder(resp.Body).Decode(&billActions)
	if err != nil {
		return LatestBillActions{}, err
	}

	return billActions, nil
}

func (c *CongGovApiClient) GetBillActions(congressNumber, billNumber int, billType string) (BillActions, error) {
	path := fmt.Sprintf("v3/bill/%d/%s/%d/actions", congressNumber, billType, billNumber)
	url := c.FormatUrl(path)

	fmt.Println(url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return BillActions{}, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return BillActions{}, err
	}
	defer resp.Body.Close()

	var billActions BillActions
	err = json.NewDecoder(resp.Body).Decode(&billActions)
	if err != nil {
		return BillActions{}, err
	}

	return billActions, nil
}
