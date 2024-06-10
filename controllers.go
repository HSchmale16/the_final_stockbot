package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetToProcess(c *gin.Context) {
	db, _ := c.MustGet("db").(*gorm.DB)

	var count int64
	db.Debug().Model(&RSSItem{}).Where("id NOT IN (SELECT rss_item_id FROM item_tag_rss_items WHERE model_id = ?)", 3).Count(&count)

	c.JSON(200, gin.H{
		"count": count,
	})
}

func GetLoadedArticlesStatus(c *gin.Context) {
	db, _ := c.MustGet("db").(*gorm.DB)

	var count int64
	db.Model(&RSSItem{}).Where("DATE(pub_date) >= DATE('now', '-2 days')").Count(&count)
	c.JSON(http.StatusOK, gin.H{
		"count": count,
	})
}

func AnalyzeTagsForReferences(c *gin.Context) {
	db, _ := c.MustGet("db").(*gorm.DB)

	// Execute arbitrary SQL code
	result := db.Exec("YOUR_SQL_CODE_HERE")

	if result.Error != nil {
		// Handle error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Analyzed tags for references",
	})
}

/**
 * Get the topics for a given time period
 * Queries the ItemTagRSSItem table to get the tags for the given time period based on the pub date of the rss items
 * Returns the counts for the tag values, and the tag name itself
 */
func GetTopicsForTimePeriod(c *gin.Context) {
	db, _ := c.MustGet("db").(*gorm.DB)
	var topics []struct {
		TagName string
		Count   int
	}
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	db.Debug().Model(&ItemTagRSSItem{}).
		Select("item_tags.name as TagName, count(*) as Count").
		Joins("JOIN rss_items ON rss_items.id = item_tag_rss_items.rss_item_id").
		Joins("JOIN item_tags ON item_tags.id = item_tag_rss_items.item_tag_id").
		Where("DATE(rss_items.pub_date) >= ?", startDate).
		Where("DATE(rss_items.pub_date) <= ?", endDate).
		Group("item_tags.name").
		Order("Count DESC").
		Having("Count > 1").
		Scan(&topics)

	c.JSON(http.StatusOK, gin.H{
		"topics": topics,
	})
}
