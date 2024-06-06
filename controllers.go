package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func GetLoadedArticlesStatus(c *gin.Context) {

	db, _ := c.MustGet("db").(*gorm.DB)

	var count int64
	db.Model(&RSSItem{}).Where("DATE(published_date) >= DATE('now', '-2 days')").Count(&count)
	c.JSON(http.StatusOK, gin.H{
		"count": count,
	})
}
