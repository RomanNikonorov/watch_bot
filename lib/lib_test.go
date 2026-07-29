package lib

import (
	"testing"
	"time"
)

func TestStartOfDayUsesCalendarDateInProvidedLocation(t *testing.T) {
	location := time.FixedZone("Asia/Novosibirsk", 7*60*60)
	value := time.Date(2026, time.July, 29, 3, 15, 42, 123, location)

	result := StartOfDay(value)
	expected := time.Date(2026, time.July, 29, 0, 0, 0, 0, location)

	if !result.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
	if result.Location() != location {
		t.Fatalf("expected location %v, got %v", location, result.Location())
	}
}

func TestStartOfDayDoesNotUseAbsoluteDurationBoundary(t *testing.T) {
	location := time.FixedZone("Asia/Novosibirsk", 7*60*60)
	value := time.Date(2026, time.July, 29, 3, 0, 0, 0, location)

	result := StartOfDay(value)
	truncated := value.Truncate(24 * time.Hour)

	if result.Day() != 29 {
		t.Fatalf("expected local calendar day 29, got %d", result.Day())
	}
	if result.Equal(truncated) {
		t.Fatalf("expected local midnight %v to differ from duration boundary %v", result, truncated)
	}
}
