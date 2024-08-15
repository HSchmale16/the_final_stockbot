package senatelobbying

import (
	"fmt"
	"io"
	"net/http"
)

const BASE_URL = "https://lda.senate.gov/api/v1/"

/*
 * Implements a very very stupid request handler to download the things.
 * It does not gracefully recover from errors, it does not handle rate limiting
 */
func SendRequest(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// This is an anon api but it's a hell of a lot faster with a token
	if senate_token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Token %s", senate_token))
	}

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
