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

func TestGetVotesForMember_Last30Days(t *testing.T) {
	app := fiber.New(fiber.Config{
		Views: m.GetTemplateEngine(),
	})

	db, _ := m.SetupDB()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("db", db)
		return c.Next()
	})
	SetupRoutes(app)

	// Clean and create test member
	testMemberId := "TEST_MEMBER_30D"
	db.Unscoped().Where("bio_guide_id = ?", testMemberId).Delete(&m.DB_CongressMember{})
	db.Create(&m.DB_CongressMember{
		BioGuideId: testMemberId,
		Name:       "Test Senator 30D",
		IsActive:   true,
	})
	defer func() {
		db.Unscoped().Where("bio_guide_id = ?", testMemberId).Delete(&m.DB_CongressMember{})
	}()

	// Create a vote 10 days ago (within 30 days) and 45 days ago (outside 30 days)
	voteRecent := Vote{
		ActionAt:    time.Now().AddDate(0, 0, -10),
		Chamber:     "Senate",
		RollCallNum: 8801,
		CongressNum: 118,
		Session:     "2",
		VoteResult:  "Passed",
		LegisName:   "S. 3001 (Recent)",
		Url:         "test-url-recent-30d",
	}
	voteOld := Vote{
		ActionAt:    time.Now().AddDate(0, 0, -45),
		Chamber:     "Senate",
		RollCallNum: 8802,
		CongressNum: 118,
		Session:     "2",
		VoteResult:  "Passed",
		LegisName:   "S. 3002 (Old)",
		Url:         "test-url-old-30d",
	}
	db.Unscoped().Where("url IN ?", []string{voteRecent.Url, voteOld.Url}).Delete(&Vote{})
	db.Create(&voteRecent)
	db.Create(&voteOld)
	defer func() {
		db.Unscoped().Where("member_id = ?", testMemberId).Delete(&VoteRecord{})
		db.Unscoped().Where("url IN ?", []string{voteRecent.Url, voteOld.Url}).Delete(&Vote{})
	}()

	db.Create(&VoteRecord{VoteID: voteRecent.ID, MemberId: testMemberId, VoteStatus: "Yea"})
	db.Create(&VoteRecord{VoteID: voteOld.ID, MemberId: testMemberId, VoteStatus: "Yea"})

	req := httptest.NewRequest(http.MethodGet, "/htmx/votes/member/"+testMemberId, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)

	if !strings.Contains(body, "Recent Votes Cast (Last 30 Days)") {
		t.Errorf("Expected body to contain 'Recent Votes Cast (Last 30 Days)', got: %s", body)
	}
	if !strings.Contains(body, "S. 3001 (Recent)") {
		t.Errorf("Expected body to contain recent vote S. 3001 (Recent)")
	}
	if strings.Contains(body, "S. 3002 (Old)") {
		t.Errorf("Expected body NOT to contain old vote S. 3002 (Old) in the recent 30-day table")
	}
}
