package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

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
