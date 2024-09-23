package congressgov

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

func (c *CongGovApiClient) GetBillsFromCongress(congressNum, offset int) (LatestBillActions, error) {
	path := fmt.Sprintf("v3/bill/%d", congressNum)
	url := c.FormatUrl(path)
	url += fmt.Sprintf("&offset=%d", offset)
	fmt.Println(url)

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

func (c *CongGovApiClient) GetBillActions(congressNumber int, billNumber, billType string) (BillActions, error) {
	path := fmt.Sprintf("v3/bill/%d/%s/%s/actions", congressNumber, strings.ToLower(billType), billNumber)
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

	if resp.StatusCode != http.StatusOK {
		return BillActions{}, fmt.Errorf("error getting bill actions: %v", resp)
	}

	var billActions BillActions
	err = json.NewDecoder(resp.Body).Decode(&billActions)
	if err != nil {
		return BillActions{}, err
	}

	return billActions, nil
}

func (c *CongGovApiClient) GetBillCosponsors(offset, congressNumber int, billNumber, billType string) (CosponsorsResponse, error) {
	path := fmt.Sprintf("v3/bill/%d/%s/%s/cosponsors", congressNumber, strings.ToLower(billType), billNumber)
	url := c.FormatUrl(path)

	fmt.Println(url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return CosponsorsResponse{}, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return CosponsorsResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return CosponsorsResponse{}, fmt.Errorf("error getting bill cosponsors: %v", resp)
	}

	var cosponsors CosponsorsResponse
	err = json.NewDecoder(resp.Body).Decode(&cosponsors)
	if err != nil {
		return CosponsorsResponse{}, err
	}

	return cosponsors, nil
}

func (c *CongGovApiClient) GetBillDetails(congressNumber int, billNumber, billType string) ([]byte, error) {
	path := fmt.Sprintf("v3/bill/%d/%s/%s", congressNumber, strings.ToLower(billType), billNumber)
	url := c.FormatUrl(path)

	fmt.Println(url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error getting bill details: %v", resp)
	}

	return io.ReadAll(resp.Body)
}

type BillDetails map[string]interface{}
