package votes

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hschmale16/the_final_stockbot/internal/m"
)

func TestRedirectToCurrentYearAttendance(t *testing.T) {
	app := fiber.New()
	SetupRoutes(app)

	currentYear := time.Now().Year()

	t.Run("Without query params", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/attendance", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if resp.StatusCode != http.StatusFound {
			t.Errorf("Expected status code %d, got %d", http.StatusFound, resp.StatusCode)
		}

		expectedLocation := fmt.Sprintf("/attendance/%d", currentYear)
		if resp.Header.Get("Location") != expectedLocation {
			t.Errorf("Expected Location header %s, got %s", expectedLocation, resp.Header.Get("Location"))
		}
	})

	t.Run("With query params", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/attendance?chamber=Senate&tenure=veteran", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if resp.StatusCode != http.StatusFound {
			t.Errorf("Expected status code %d, got %d", http.StatusFound, resp.StatusCode)
		}

		expectedLocation := fmt.Sprintf("/attendance/%d?chamber=Senate&tenure=veteran", currentYear)
		if resp.Header.Get("Location") != expectedLocation {
			t.Errorf("Expected Location header %s, got %s", expectedLocation, resp.Header.Get("Location"))
		}
	})
}

func TestHtmxAttendanceReplaceUrlHeaderAndFilterNotice(t *testing.T) {
	app := fiber.New(fiber.Config{
		Views: m.GetTemplateEngine(),
	})

	db, _ := m.SetupDB()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("db", db)
		return c.Next()
	})
	SetupRoutes(app)

	t.Run("With query parameters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/htmx/attendance/2026?chamber=Senate&tenure=newcomer&age=under50", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expectedReplaceURL := "/attendance/2026?chamber=Senate&tenure=newcomer&age=under50"
		if got := resp.Header.Get("HX-Replace-Url"); got != expectedReplaceURL {
			t.Errorf("Expected HX-Replace-Url %q, got %q", expectedReplaceURL, got)
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		body := string(bodyBytes)
		if !strings.Contains(body, "Table is shorter due to applied filters") {
			t.Errorf("Expected body to contain filter notice, got: %s", body)
		}
	})

	t.Run("Without query parameters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/htmx/attendance/2026", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expectedReplaceURL := "/attendance/2026"
		if got := resp.Header.Get("HX-Replace-Url"); got != expectedReplaceURL {
			t.Errorf("Expected HX-Replace-Url %q, got %q", expectedReplaceURL, got)
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		body := string(bodyBytes)
		if strings.Contains(body, "Table is shorter due to applied filters") {
			t.Errorf("Expected body not to contain filter notice, got: %s", body)
		}
	})
}

func TestTermForYearAndWasInOfficeOn(t *testing.T) {
	leg := m.US_CongressLegislator{
		Terms: []m.Terms{
			{
				Type:  "rep",
				State: "CA",
				Start: "2019-01-03",
				End:   "2023-01-03",
				Party: "Democrat",
			},
			{
				Type:  "sen",
				State: "CA",
				Start: "2023-01-03",
				End:   "2029-01-03",
				Party: "Democrat",
			},
		},
	}

	// 2021 was Rep
	term2021, ok := leg.TermForYear(2021)
	if !ok {
		t.Fatalf("Expected TermForYear(2021) to return true")
	}
	if term2021.Type != "rep" {
		t.Errorf("Expected term type 'rep' for 2021, got '%s'", term2021.Type)
	}

	// 2024 was Sen
	term2024, ok := leg.TermForYear(2024)
	if !ok {
		t.Fatalf("Expected TermForYear(2024) to return true")
	}
	if term2024.Type != "sen" {
		t.Errorf("Expected term type 'sen' for 2024, got '%s'", term2024.Type)
	}

	// 2015 was not in office
	_, ok = leg.TermForYear(2015)
	if ok {
		t.Errorf("Expected TermForYear(2015) to return false")
	}

	// WasInOfficeOn
	t1 := time.Date(2021, 6, 15, 12, 0, 0, 0, time.UTC)
	if !leg.WasInOfficeOn(t1) {
		t.Errorf("Expected WasInOfficeOn to be true on %v", t1)
	}

	t2 := time.Date(2018, 12, 31, 0, 0, 0, 0, time.UTC)
	if leg.WasInOfficeOn(t2) {
		t.Errorf("Expected WasInOfficeOn to be false on %v", t2)
	}
}
