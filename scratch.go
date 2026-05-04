package main

import (
	"fmt"

	"github.com/hschmale16/the_final_stockbot/internal/congress"
	"github.com/hschmale16/the_final_stockbot/internal/m"
)

func turd() {
	db, _ := m.SetupDB()
	var h congress.Hearing
	db.Preload("AttendedMembers").Where("id = ?", 72).First(&h) // 72 is an expanded hearing ID from the logs
	for _, member := range h.AttendedMembers {
		fmt.Printf("Member: %s, Party: '%s'\n", member.Name, member.Party)
	}
}
