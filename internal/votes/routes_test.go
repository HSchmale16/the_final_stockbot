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
