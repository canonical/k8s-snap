package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// SecondsToExpirationDate returns the expiration date from now plus the given
// number of seconds.
func SecondsToExpirationDate(now time.Time, seconds int) time.Time {
	switch {
	case seconds%(365*24*60*60) == 0:
		return now.AddDate(seconds/365/24/60/60, 0, 0)
	case seconds%(30*24*60*60) == 0:
		return now.AddDate(0, seconds/30/24/60/60, 0)
	case seconds%(24*60*60) == 0:
		return now.AddDate(0, 0, seconds/24/60/60)
	default:
		return now.Add(time.Duration(seconds) * time.Second)
	}
}

// TTLToSeconds converts a TTL string to seconds. The TTL string is a number
// followed by a unit:
// - y: years
// - m: months
// - d: days
func TTLToSeconds(ttl string) (int, error) {
	if len(ttl) < 2 {
		return 0, fmt.Errorf("invalid TTL format: %s", ttl)
	}

	value, err := strconv.Atoi(ttl[:len(ttl)-1])
	if err != nil {
		return 0, fmt.Errorf("invalid TTL number: %w", err)
	}

	unit := strings.ToLower(ttl[len(ttl)-1:])
	switch unit {
	case "y":
		return value * 365 * 24 * 60 * 60, nil
	case "m":
		return value * 30 * 24 * 60 * 60, nil
	case "d":
		return value * 24 * 60 * 60, nil
	default:
		return 0, fmt.Errorf("invalid TTL unit: %s", unit)
	}
}
