package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSearch(t *testing.T) {
	// Create a mock database connection
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	// Create a new Gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("db", db)

	// Call the Search function
	Search(c)

	// Check the response status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, w.Code)
	}

	// Parse the response body
	var tags []Tag
	err := json.Unmarshal(w.Body.Bytes(), &tags)
	if err != nil {
		t.Errorf("Failed to parse response body: %v", err)
	}

	// Add your assertions here
	// ...

}
