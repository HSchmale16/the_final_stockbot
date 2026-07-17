package main

import (
	"fmt"

	"github.com/hschmale16/the_final_stockbot/internal/m"
)

func turd() {
	db, _ := m.SetupDB()
	var runs []m.CronJobRun
	db.Order("created_at DESC").Find(&runs)
	fmt.Printf("Total runs: %d\n", len(runs))
	for _, run := range runs {
		fmt.Printf("Job: %s | Status: %s | Items: %d | Message: %s\n", run.JobName, run.Status, run.ItemsFound, run.Message)
	}
}
