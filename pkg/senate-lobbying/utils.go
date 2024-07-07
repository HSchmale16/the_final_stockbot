package senatelobbying

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

var SENATE_TOKEN string

func init() {
	SENATE_TOKEN = os.Getenv("SENATE_TOKEN")
	fmt.Println("SENATE TOKEN:", SENATE_TOKEN)
}

/*
 * Implements a very very stupid request handler to download the things.
 * It does not gracefully recover from errors, it does not handle rate limiting
 */
func SendRequest(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", SENATE_TOKEN))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("retry %s", resp.Header.Get("Retry-After"))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
