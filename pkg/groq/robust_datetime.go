package henry_groq

import (
	"errors"
	"time"
)

// Define common datetime layouts
var datetimeLayouts = []string{
	time.RFC822,                       // "02 Jan 06 15:04 MST"
	time.RFC822Z,                      // "02 Jan 06 15:04 -0700"
	time.RFC1123,                      // "Mon, 02 Jan 2006 15:04:05 MST"
	time.RFC1123Z,                     // "Mon, 02 Jan 2006 15:04:05 -0700"
	time.RFC3339,                      // "2006-01-02T15:04:05Z07:00"
	"2006-01-02T15:04:05.000Z07:00",   // ISO8601 with milliseconds
	"Mon, 02 Jan 2006 15:04:05 -0700", // RSS 2.0 spec
	"Mon, 02 Jan 2006 15:04:05 MST",   // RSS 2.0 spec alternative
	"2006-01-02",
	"01/02/2006",
}

// parseDatetime attempts to parse a datetime string using known layouts
func ParseDateTimeRssRobustly(datetimeStr string) (time.Time, error) {
	var parsedTime time.Time
	var err error

	for _, layout := range datetimeLayouts {
		parsedTime, err = time.Parse(layout, datetimeStr)
		if err == nil {
			return parsedTime, nil
		}
	}

	return time.Time{}, errors.New("unable to parse datetime string")
}
