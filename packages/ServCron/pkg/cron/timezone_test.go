package cron

import (
	"testing"
	"time"
)

func TestTimezoneScheduler_CalculateNextCronTZ(t *testing.T) {
	tzs := NewTimezoneScheduler()

	refTime := time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC)

	// Every day at 09:00 in America/New_York (UTC-4 in July)
	nextNY, err := tzs.CalculateNextCronTZ("0 9 * * *", "America/New_York", refTime)
	if err != nil {
		t.Fatalf("CalculateNextCronTZ failed: %v", err)
	}

	// 09:00 EDT on July 26 is 13:00 UTC (06:00 EDT < 09:00 EDT)
	expectedUTC := time.Date(2026, 7, 26, 13, 0, 0, 0, time.UTC)
	if !nextNY.Equal(expectedUTC) {
		t.Errorf("expected next NY run at %v, got %v", expectedUTC, nextNY.UTC())
	}

	// Test invalid IANA location
	_, err = tzs.CalculateNextCronTZ("0 9 * * *", "Invalid/Timezone_Location", refTime)
	if err == nil {
		t.Error("expected error for invalid location name")
	}
}

func TestFormatTimeInZone(t *testing.T) {
	refTime := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

	formatted, err := FormatTimeInZone(refTime, "Asia/Tokyo", "2006-01-02 15:04:05 MST")
	if err != nil {
		t.Fatalf("FormatTimeInZone failed: %v", err)
	}

	// UTC 12:00 is Tokyo 21:00 JST
	expected := "2026-07-26 21:00:00 JST"
	if formatted != expected {
		t.Errorf("expected Tokyo time %s, got %s", expected, formatted)
	}
}
