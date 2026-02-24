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

type TripInfo struct {
	ID            uint
	AggregatedId  string
	Destination   string
	DepartureDate time.Time
	ReturnDate    time.Time
	TravelSponsor string
	Duration      int
	Color         string
	IsFirstDay    bool
	IsLastDay     bool
	Lane          int
	TopOffset     string
	Count         int
}

type DatesStruct []struct {
	MonthNumber int
	DayNumber   int
	IsToday     bool
	Trips       []TripInfo
}

type AggregatedTrip struct {
	ID            string
	Destination   string
	DepartureDate time.Time
	ReturnDate    time.Time
	Count         int
}

var colors = []string{"#e6194B", "#3cb44b", "#ffe119", "#4363d8", "#f58231", "#911eb4", "#42d4f4", "#f032e6", "#bfef45", "#fabed4", "#469990", "#dcbeff", "#9A6324", "#fffac8", "#800000", "#aaffc3", "#808000", "#ffd8b1", "#000075", "#a9a9a9"}
var tripColors = make(map[string]string)
var colorIndex = 0

func assignColorToTrip(tripId string) string {
	if color, exists := tripColors[tripId]; exists {
		return color
	}
	color := colors[colorIndex]
	tripColors[tripId] = color
	colorIndex = (colorIndex + 1) % len(colors)
	return color
}

func GetTravelCalendarData(year, month int, db *gorm.DB) []TravelWeek {
	colorIndex = 0 // Reset for each call
	tripColors = make(map[string]string)

	firstDayOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	lastDayOfMonth := firstDayOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

	var travels []DB_TravelDisclosure
	db.Debug().
		Where("departure_date <= ? AND return_date >= ?", lastDayOfMonth, firstDayOfMonth).
		Order("departure_date").
		Find(&travels)

	// Aggregate trips
	aggregatedTripsMap := make(map[string]*AggregatedTrip)
	aggregatedTrips := make([]*AggregatedTrip, 0)
	for _, t := range travels {
		key := fmt.Sprintf("%s-%s-%s", t.Destination, t.DepartureDate.Format(time.RFC3339), t.ReturnDate.Format(time.RFC3339))
		if aggTrip, exists := aggregatedTripsMap[key]; exists {
			aggTrip.Count++
		} else {
			newAggTrip := &AggregatedTrip{
				ID:            key,
				Destination:   t.Destination,
				DepartureDate: t.DepartureDate,
				ReturnDate:    t.ReturnDate,
				Count:         1,
			}
			aggregatedTripsMap[key] = newAggTrip
			aggregatedTrips = append(aggregatedTrips, newAggTrip)
		}
	}
	sort.Slice(aggregatedTrips, func(i, j int) bool {
		return aggregatedTrips[i].DepartureDate.Before(aggregatedTrips[j].DepartureDate)
	})

	// Assign lanes to trips
	lanes := make([][]*AggregatedTrip, 0)
	tripLane := make(map[string]int)
	for _, trip := range aggregatedTrips {
		placed := false
		for i, lane := range lanes {
			overlap := false
			for _, placedTrip := range lane {
				if trip.DepartureDate.Before(placedTrip.ReturnDate) && placedTrip.DepartureDate.Before(trip.ReturnDate) {
					overlap = true
					break
				}
			}
			if !overlap {
				lanes[i] = append(lanes[i], trip)
				tripLane[trip.ID] = i
				placed = true
				break
			}
		}
		if !placed {
			lanes = append(lanes, []*AggregatedTrip{trip})
			tripLane[trip.ID] = len(lanes) - 1
		}
	}

	weeks := make([]TravelWeek, 0)
	currentDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	for currentDate.Weekday() != time.Sunday {
		currentDate = currentDate.AddDate(0, 0, -1)
	}

	for weekNum := 0; ; weekNum++ {
		if currentDate.Month() != time.Month(month) && currentDate.Day() > 7 {
			if currentDate.Month() > time.Month(month) || currentDate.Year() > year {
				allNextMonth := true
				for i := 0; i < 7; i++ {
					if currentDate.AddDate(0, 0, i).Month() == time.Month(month) {
						allNextMonth = false
						break
					}
				}
				if allNextMonth {
					break
				}
			}
		}
		if weekNum > 6 {
			break
		}

		week := TravelWeek{
			WeekNumber: weekNum,
			Dates:      make(DatesStruct, 7),
		}

		for dayOfWeek := 0; dayOfWeek < 7; dayOfWeek++ {
			day := currentDate
			week.Dates[dayOfWeek].DayNumber = day.Day()
			week.Dates[dayOfWeek].MonthNumber = int(day.Month())
			week.Dates[dayOfWeek].IsToday = day.Year() == time.Now().Year() && day.Month() == time.Now().Month() && day.Day() == time.Now().Day()

			dayStart := day
			dayEnd := day.Add(24 * time.Hour)

			for _, t := range aggregatedTrips {
				if t.DepartureDate.Before(dayEnd) && t.ReturnDate.After(dayStart) {
					if week.Dates[dayOfWeek].Trips == nil {
						week.Dates[dayOfWeek].Trips = make([]TripInfo, 0)
					}
					duration := int(t.ReturnDate.Sub(t.DepartureDate).Hours()/24) + 1
					isFirstDay := t.DepartureDate.Year() == day.Year() && t.DepartureDate.Month() == day.Month() && t.DepartureDate.Day() == day.Day()
					isLastDay := t.ReturnDate.Year() == day.Year() && t.ReturnDate.Month() == day.Month() && t.ReturnDate.Day() == day.Day()

					lane := tripLane[t.ID]
					topOffset := fmt.Sprintf("%.2frem", 1.5+float64(lane)*1.75)

					destination := t.Destination
					if t.Count > 1 {
						destination = destination + " (" + strconv.Itoa(t.Count) + ")"
					}

					week.Dates[dayOfWeek].Trips = append(week.Dates[dayOfWeek].Trips, TripInfo{
						AggregatedId:  t.ID,
						Destination:   destination,
						DepartureDate: t.DepartureDate,
						ReturnDate:    t.ReturnDate,
						Duration:      duration,
						Color:         assignColorToTrip(t.ID),
						IsFirstDay:    isFirstDay,
						IsLastDay:     isLastDay,
						Lane:          lane,
						TopOffset:     topOffset,
						Count:         t.Count,
					})
				}
			}
			sort.Slice(week.Dates[dayOfWeek].Trips, func(i, j int) bool {
				return week.Dates[dayOfWeek].Trips[i].Lane < week.Dates[dayOfWeek].Trips[j].Lane
			})
			currentDate = currentDate.AddDate(0, 0, 1)
		}

		weeks = append(weeks, week)
	}

	return weeks
}
