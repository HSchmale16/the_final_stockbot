package votes

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hschmale16/the_final_stockbot/internal/m"
)

func TestComputeYearVoteFrequency_Structure(t *testing.T) {
	db, err := m.SetupDB()
	if err != nil {
		t.Fatalf("Failed to setup DB: %v", err)
	}

	// 1999 has no votes in db
	freq := computeYearVoteFrequency(db, 1999, "All")
	if freq.Year != 1999 {
		t.Errorf("Expected year 1999, got %d", freq.Year)
	}
	if len(freq.Months) != 12 {
		t.Fatalf("Expected 12 months, got %d", len(freq.Months))
	}

	// Verify month day counts (2024 leap year: Feb has 29 days)
	freq2024 := computeYearVoteFrequency(db, 2024, "All")
	expectedDays2024 := []int{31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	for i, mon := range freq2024.Months {
		if len(mon.Days) != expectedDays2024[i] {
			t.Errorf("Month %d (%s) expected %d days, got %d", mon.MonthNumber, mon.MonthName, expectedDays2024[i], len(mon.Days))
		}
	}

	// Verify non-leap year (2023 Feb has 28 days)
	freq2023 := computeYearVoteFrequency(db, 2023, "All")
	if len(freq2023.Months[1].Days) != 28 {
		t.Errorf("2023 Feb expected 28 days, got %d", len(freq2023.Months[1].Days))
	}
}

func TestComputeYearVoteFrequency_Calculations(t *testing.T) {
	db, err := m.SetupDB()
	if err != nil {
		t.Fatalf("Failed to setup DB: %v", err)
	}

	// Insert temporary test votes for year 1998
	testYear := 1998
	db.Unscoped().Where("action_at BETWEEN ? AND ?", time.Date(testYear, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(testYear, 12, 31, 23, 59, 59, 0, time.UTC)).Delete(&Vote{})

	testVotes := []Vote{
		{ActionAt: time.Date(testYear, 1, 15, 14, 0, 0, 0, time.UTC), Chamber: "House", RollCallNum: 9001, CongressNum: 105, Session: "2", VoteResult: "Passed", Url: "test-url-1"},
		{ActionAt: time.Date(testYear, 1, 15, 15, 0, 0, 0, time.UTC), Chamber: "House", RollCallNum: 9002, CongressNum: 105, Session: "2", VoteResult: "Passed", Url: "test-url-2"},
		{ActionAt: time.Date(testYear, 1, 15, 16, 0, 0, 0, time.UTC), Chamber: "Senate", RollCallNum: 9003, CongressNum: 105, Session: "2", VoteResult: "Failed", Url: "test-url-3"},
		{ActionAt: time.Date(testYear, 3, 20, 10, 0, 0, 0, time.UTC), Chamber: "House", RollCallNum: 9004, CongressNum: 105, Session: "2", VoteResult: "Passed", Url: "test-url-4"},
		{ActionAt: time.Date(testYear, 3, 20, 11, 0, 0, 0, time.UTC), Chamber: "House", RollCallNum: 9005, CongressNum: 105, Session: "2", VoteResult: "Passed", Url: "test-url-5"},
		{ActionAt: time.Date(testYear, 3, 20, 12, 0, 0, 0, time.UTC), Chamber: "House", RollCallNum: 9006, CongressNum: 105, Session: "2", VoteResult: "Passed", Url: "test-url-6"},
	}
	for _, v := range testVotes {
		if err := db.Create(&v).Error; err != nil {
			t.Fatalf("Failed to insert test vote: %v", err)
		}
	}
	defer func() {
		db.Unscoped().Where("action_at BETWEEN ? AND ?", time.Date(testYear, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(testYear, 12, 31, 23, 59, 59, 0, time.UTC)).Delete(&Vote{})
	}()

	freqAll := computeYearVoteFrequency(db, testYear, "All")
	if freqAll.TotalVotes != 6 {
		t.Errorf("Expected 6 total votes, got %d", freqAll.TotalVotes)
	}
	if freqAll.VotingDays != 2 {
		t.Errorf("Expected 2 voting days, got %d", freqAll.VotingDays)
	}
	if freqAll.MaxDaily != 3 {
		t.Errorf("Expected max daily 3, got %d", freqAll.MaxDaily)
	}
	if freqAll.BusiestDate != "Jan 15" && freqAll.BusiestDate != "Mar 20" {
		t.Errorf("Expected busiest date 'Jan 15' or 'Mar 20', got '%s'", freqAll.BusiestDate)
	}

	// Jan checks
	jan := freqAll.Months[0]
	if jan.TotalVotes != 3 {
		t.Errorf("Expected Jan total votes 3, got %d", jan.TotalVotes)
	}
	jan15 := jan.Days[14]
	if jan15.DayNumber != 15 || jan15.Count != 3 || !jan15.HasVotes {
		t.Errorf("Expected Jan 15 to have 3 votes, got %+v", jan15)
	}
	if jan15.HeightPct != 100 {
		t.Errorf("Expected height 100%%, got %d%%", jan15.HeightPct)
	}
	if !strings.Contains(jan15.Tooltip, "3 votes") {
		t.Errorf("Expected tooltip to contain '3 votes', got '%s'", jan15.Tooltip)
	}

	// Chamber filtering
	freqHouse := computeYearVoteFrequency(db, testYear, "House")
	if freqHouse.TotalVotes != 5 {
		t.Errorf("Expected House total votes 5, got %d", freqHouse.TotalVotes)
	}
	freqSenate := computeYearVoteFrequency(db, testYear, "Senate")
	if freqSenate.TotalVotes != 1 {
		t.Errorf("Expected Senate total votes 1, got %d", freqSenate.TotalVotes)
	}
}

func TestGetAttendanceYearPage_RendersSparkline(t *testing.T) {
	app := fiber.New(fiber.Config{
		Views: m.GetTemplateEngine(),
	})

	db, _ := m.SetupDB()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("db", db)
		return c.Next()
	})
	SetupRoutes(app)

	req := httptest.NewRequest(http.MethodGet, "/attendance/2026", nil)
	resp, err := app.Test(req, 10000)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)

	if !strings.Contains(body, "Annual Voting Frequency") {
		t.Errorf("Expected rendered page to contain 'Annual Voting Frequency', got: %s", body)
	}
	if !strings.Contains(body, "Jan") || !strings.Contains(body, "Dec") {
		t.Errorf("Expected rendered page to contain month segments Jan..Dec")
	}
}
