package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// YearSeconds is the number of seconds in a year.
	// 365 days * 24 hours * 60 minutes * 60 seconds. This is an approximation.
	yearSeconds = 365 * 24 * 60 * 60
	// MonthSeconds is the number of seconds in a month.
	// 30 days * 24 hours * 60 minutes * 60 seconds. This is an approximation.
	monthSeconds = 30 * 24 * 60 * 60
	// DaySeconds is the number of seconds in a day.
	// 24 hours * 60 minutes * 60 seconds. This is an approximation.
	daySeconds = 24 * 60 * 60
)

// SecondsToExpirationDate returns the expiration date from now plus the given
// number of seconds.
func SecondsToExpirationDate(now time.Time, seconds int) time.Time {
	switch {
	case seconds%(yearSeconds) == 0:
		return now.AddDate(seconds/yearSeconds, 0, 0)
	case seconds%(monthSeconds) == 0:
		return now.AddDate(0, seconds/monthSeconds, 0)
	case seconds%(daySeconds) == 0:
		return now.AddDate(0, 0, seconds/daySeconds)
	default:
		return now.Add(time.Duration(seconds) * time.Second)
	}
}

// TTLToSeconds converts a TTL string to seconds. The TTL string is a number
// followed by a unit:
// - y: years
// - mo: months
// - d: days
// - any other unit supported by time.ParseDuration.
func TTLToSeconds(ttl string) (int, error) {
	if len(ttl) < 2 {
		return 0, fmt.Errorf("invalid TTL length: %s", ttl)
	}

	var value int
	var unit string
	var err error

	if strings.HasSuffix(ttl, "mo") {
		value, err = strconv.Atoi(ttl[:len(ttl)-2])
		unit = "mo"
	} else {
		value, err = strconv.Atoi(ttl[:len(ttl)-1])
		unit = strings.ToLower(ttl[len(ttl)-1:])
	}

	if err != nil {
		return 0, fmt.Errorf("invalid TTL format: %w", err)
	}

	switch unit {
	case "y":
		return value * yearSeconds, nil
	case "mo":
		return value * monthSeconds, nil
	case "d":
		return value * daySeconds, nil
	default:
		duration, err := time.ParseDuration(ttl)
		if err != nil {
			return 0, fmt.Errorf("invalid TTL unit: %s", unit)
		}
		return int(duration.Seconds()), nil
	}
}
