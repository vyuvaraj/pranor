package import (
	"fmt"
	"strings"
	"time"
)

// TimezoneScheduler computes cron schedules in specific IANA timezones (e.g., "America/New_York", "Asia/Tokyo").
type TimezoneScheduler struct{}

// NewTimezoneScheduler creates a TimezoneScheduler instance.
func NewTimezoneScheduler() *TimezoneScheduler {
	return &TimezoneScheduler{}
}

// CalculateNextCronTZ evaluates a 5-field cron expression against a specified IANA timezone string.
// Handles daylight saving time (DST) offsets dynamically.
func (tzs *TimezoneScheduler) CalculateNextCronTZ(cronExpr string, tzName string, from time.Time) (time.Time, error) {
	if tzName == "" || strings.EqualFold(tzName, "UTC") {
		tzName = "UTC"
	}

	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid IANA timezone '%s': %w", tzName, err)
	}

	// Convert starting time to local target timezone
	localFrom := from.In(loc)

	// Calculate next cron run using localized reference
	nextLocal, err := CalculateNextCron(cronExpr, localFrom)
	if err != nil {
		return time.Time{}, err
	}

	return nextLocal.In(loc), nil
}

// FormatTimeInZone returns formatted time string in specified timezone.
func FormatTimeInZone(t time.Time, tzName string, layout string) (string, error) {
	if tzName == "" {
		tzName = "UTC"
	}
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return "", err
	}
	if layout == "" {
		layout = time.RFC3339
	}
	return t.In(loc).Format(layout), nil
}
