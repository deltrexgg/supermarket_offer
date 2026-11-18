// Package dubai holds the one definition of "today" for the platform.
// Offer dates are Dubai calendar days, so every date decision goes through here.
package dubai

import (
	"time"
	_ "time/tzdata" // the slim runtime image has no tz database; embed it
)

const DateLayout = "2006-01-02"

var loc = mustLoad()

func mustLoad() *time.Location {
	l, err := time.LoadLocation("Asia/Dubai")
	if err != nil {
		panic("asia/dubai timezone: " + err.Error())
	}
	return l
}

func Location() *time.Location { return loc }

func Now() time.Time { return time.Now().In(loc) }

// Today returns the Dubai calendar date as YYYY-MM-DD.
func Today() string { return Now().Format(DateLayout) }

// DateOffset returns the Dubai date n days from today as YYYY-MM-DD.
func DateOffset(n int) string { return Now().AddDate(0, 0, n).Format(DateLayout) }

// UntilMidnight is the time left until the next Dubai midnight (cache TTL).
func UntilMidnight() time.Duration {
	n := Now()
	next := time.Date(n.Year(), n.Month(), n.Day()+1, 0, 0, 0, 0, loc)
	return next.Sub(n)
}

// ValidDate reports whether s is a real YYYY-MM-DD date.
func ValidDate(s string) bool {
	_, err := time.ParseInLocation(DateLayout, s, loc)
	return err == nil
}
