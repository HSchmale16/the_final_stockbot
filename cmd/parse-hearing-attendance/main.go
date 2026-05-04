package main

import (
	"log"

	"github.com/hschmale16/the_final_stockbot/internal/app"
	"github.com/hschmale16/the_final_stockbot/internal/m"
)

func main() {
	db, err := m.SetupDB()
	if err != nil {
		log.Fatalf("Failed to setup database: %v\n", err)
	}

	log.Println("Starting Hearing Attendance parsing job...")
	
	err = app.ProcessHearingAttendance(db)
	if err != nil {
		log.Fatalf("Error processing hearing attendance: %v\n", err)
	}

	log.Println("Hearing Attendance parsing job completed successfully.")
}
