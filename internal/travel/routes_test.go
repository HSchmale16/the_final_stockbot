package travel

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestRedirectToCurrentMonthCalendar(t *testing.T) {
	app := fiber.New()
	SetupRoutes(app)

	req := httptest.NewRequest(http.MethodGet, "/travel/calendar", nil)
	resp, _ := app.Test(req, 1) // The 1s timeout is arbitrary

	// Assertions
	if resp.StatusCode != http.StatusFound {
		t.Errorf("Expected status code %d, got %d", http.StatusFound, resp.StatusCode)
	}

	expectedMonth := time.Now().AddDate(0, -1, 0)
	expectedLocation := fmt.Sprintf("/travel/calendar/%d/%d", expectedMonth.Year(), int(expectedMonth.Month()))

	if resp.Header.Get("Location") != expectedLocation {
		t.Errorf("Expected Location header %s, got %s", expectedLocation, resp.Header.Get("Location"))
	}
}