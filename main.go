package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	// Create a new Gin router
	router := gin.Default()

	// Define your routes here

	// Start the server
	router.Run(":8080")
}