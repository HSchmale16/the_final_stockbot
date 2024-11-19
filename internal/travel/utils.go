package travel

import (
	"fmt"
	"sort"
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
	MonthNumber int
	DayNumber   int
	// Destinations to number of congress critters there
	Destinations       map[string]int
	DestinationsSorted []string
}

func (d DatesStruct) MakeDestinationsSorted() {
	for i := range d {
		d[i].DestinationsSorted = make([]string, 0, len(d[i].Destinations))
		for k := range d[i].Destinations {
			dest := d[i].Destinations[k]
			destStr := k + " (" + strconv.Itoa(dest) + ")"
			d[i].DestinationsSorted = append(d[i].DestinationsSorted, destStr)
		}
		sort.Strings(d[i].DestinationsSorted)
	}
}

func GetTravelCalendarData(year, month int, db *gorm.DB) []TravelWeek {
	yearStr := strconv.Itoa(year)
	monthStr := strconv.Itoa(month)

	dateStr := fmt.Sprintf("%s-%s-01", yearStr, monthStr)

	var travels []DB_TravelDisclosure
	db.Debug().
		Where("?::date in (date_trunc('month', departure_date)::date, date_trunc('month', return_date)::date)", dateStr).
		Find(&travels)

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
			week.Dates[j].MonthNumber = int(day.Month())

			for _, t := range travels {
				if t.DepartureDate.Day() <= day.Day() && t.ReturnDate.Day() >= day.Day() {
					if week.Dates[j].Destinations == nil {
						week.Dates[j].Destinations = make(map[string]int)
					}
					week.Dates[j].Destinations[t.Destination]++
				}
			}
		}
		week.Dates.MakeDestinationsSorted()

		weeks[i] = week
	}

	return weeks
}
