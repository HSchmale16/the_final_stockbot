package main

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateTagRelationsForModel(t *testing.T) {
	// Create a mock DB connection
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open DB connection: %v", err)
	}

	// Create a mock RSS item
	item := RSSItem{
		Title:       "Test Title",
		Description: "Test Description",
	}
	db.Create(&item)

	// Create a mock LLM model
	model := LLMModel{
		ModelName: "Test Model",
	}
	db.Create(&model)

	// Create some mock tags
	tags := []string{"tag1", "tag2", "tag3"}

	// Call the function being tested
	CreateTagRelationsForModel(db, item, model, tags)

	// Verify that the tag relations were created
	var itemTags []ItemTagRSSItem
	db.Where("rss_item_id = ?", item.ID).Find(&itemTags)
	if len(itemTags) != len(tags) {
		t.Errorf("Expected %d tag relations, but got %d", len(tags), len(itemTags))
	}
}
