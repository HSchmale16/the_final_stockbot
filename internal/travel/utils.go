package travel

import (
	"strconv"
	"time"

	"gorm.io/gorm"
)

type TravelWeek struct {
	WeekNumber int
	// Dates will always have seven elements
	Dates DatesStruct
}

type DatesStruct []struct {
	DayNumber int
	// Destinations to number of congress critters there
	Destinations       map[string]int
	DestinationsSorted []string
}

func GetTravelCalendarData(year, month int, db *gorm.DB) []TravelWeek {
	yearStr := strconv.Itoa(year)
	monthStr := strconv.Itoa(month)

	var travels []DB_TravelDisclosure
	db.Debug().Where("year = ? AND (EXTRACT(MONTH FROM departure_date) = ? OR EXTRACT(MONTH FROM return_date) = ?)", yearStr, monthStr, monthStr).Find(&travels)

	weeks := make([]TravelWeek, 5)
	MAX_WEEKS_IN_MONTH := 5

	// Get the first sunday of the current or previous month
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	for firstDay.Weekday() != time.Sunday {
		firstDay = firstDay.AddDate(0, 0, -1)
	}

	// iterate by week
	for i := 0; i < MAX_WEEKS_IN_MONTH; i++ {
		week := TravelWeek{
			WeekNumber: i,
			Dates:      make(DatesStruct, 7),
		}

		// iterate by day
		for j := 0; j < 7; j++ {
			day := firstDay.AddDate(0, 0, i*7+j)
			week.Dates[j].DayNumber = day.Day()

			for _, t := range travels {
				if t.DepartureDate.Day() <= day.Day() && t.ReturnDate.Day() >= day.Day() {
					if week.Dates[j].Destinations == nil {
						week.Dates[j].Destinations = make(map[string]int)
					}
					week.Dates[j].Destinations[t.Destination]++
				}
			}
		}

		weeks[i] = week
	}

	return weeks
}
