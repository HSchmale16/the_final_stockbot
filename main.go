package main

import (
	"log"
	"github.com/robfig/cron"
)


func main() {
	db, err := setupDB()
	if err != nil {
		// Handle error
		log.Println("Error occurred:", err)
		return
	}
	defer db.Close()

	c := cron.New()
	c.AddFunc("@every 15m", func() {

	})
	c.Start()

	fetchFeeds(db)
}

