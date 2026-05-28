package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestOpenGraphTags(t *testing.T) {
	// Ensure we are in a clean state and have access to templates
	os.Setenv("DEBUG", "true") // This allows using local static files and templates if needed
	
	app := SetupServer()

	tests := []struct {
		name string
		url  string
	}{
		{"Home", "/"},
		{"Tags List", "/tags"},
		{"Tag View", "/tag/1"},
		{"Law Index", "/laws"},
		{"Law View", "/law/1"},
		{"Congress Members", "/congress-members"},
		{"Congress Member View", "/congress-member/B000944"},
		{"Hearings", "/hearings"},
		{"TOS", "/tos"},
		{"Congress Network", "/congress-network"},
	}

	requiredTags := []string{
		"og:title",
		"og:description",
		"og:url",
		"og:site_name",
		"og:type",
		"og:image",
		"twitter:card",
		"twitter:title",
		"twitter:description",
		"twitter:image",
		"link rel=\"canonical\"",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			resp, err := app.Test(req, 10000) // 10s timeout
			if err != nil {
				t.Fatalf("Failed to test route %s: %v", tt.url, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status OK for %s, got %d", tt.url, resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read body for %s: %v", tt.url, err)
			}

			bodyStr := string(body)
			for _, tag := range requiredTags {
				if !strings.Contains(bodyStr, tag) {
					t.Errorf("Route %s missing required tag: %s", tt.url, tag)
				}
			}

			if tt.name == "Congress Member View" {
				if !strings.Contains(bodyStr, "static/img/muddy-") {
					t.Errorf("Congress Member View missing party-specific OgImage")
				}
			}
		})
	}
}
