package duty

import (
	"testing"
	"time"

	"watch_bot/dao"
)

func TestFindCurrentDuty_EmptyList(t *testing.T) {
	result := FindCurrentDuty([]dao.Duty{}, time.Now())
	if result != nil {
		t.Errorf("expected nil for empty list, got %+v", result)
	}
}

func TestFindCurrentDuty_TodayExists(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)

	duties := []dao.Duty{
		{ID: 1, DutyID: "alice", LastDutyDate: &yesterday},
		{ID: 2, DutyID: "bob", LastDutyDate: &today},
		{ID: 3, DutyID: "charlie", LastDutyDate: nil},
	}

	result := FindCurrentDuty(duties, today)
	if result == nil {
		t.Fatal("expected duty, got nil")
	}
	if result.DutyID != "bob" {
		t.Errorf("expected bob (today's duty), got %s", result.DutyID)
	}
}

func TestFindCurrentDuty_NextAlphabetically(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)

	duties := []dao.Duty{
		{ID: 1, DutyID: "alice", LastDutyDate: nil},
		{ID: 2, DutyID: "bob", LastDutyDate: &yesterday},
		{ID: 3, DutyID: "charlie", LastDutyDate: nil},
	}

	result := FindCurrentDuty(duties, today)
	if result == nil {
		t.Fatal("expected duty, got nil")
	}
	if result.DutyID != "charlie" {
		t.Errorf("expected charlie (next after bob), got %s", result.DutyID)
	}
}

func TestFindCurrentDuty_WrapAround(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)

	duties := []dao.Duty{
		{ID: 1, DutyID: "alice", LastDutyDate: nil},
		{ID: 2, DutyID: "bob", LastDutyDate: nil},
		{ID: 3, DutyID: "charlie", LastDutyDate: &yesterday},
	}

	result := FindCurrentDuty(duties, today)
	if result == nil {
		t.Fatal("expected duty, got nil")
	}
	if result.DutyID != "alice" {
		t.Errorf("expected alice (wrap around to first), got %s", result.DutyID)
	}
}

func TestFindCurrentDuty_NoOneHasBeenOnDuty(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)

	duties := []dao.Duty{
		{ID: 1, DutyID: "charlie", LastDutyDate: nil},
		{ID: 2, DutyID: "alice", LastDutyDate: nil},
		{ID: 3, DutyID: "bob", LastDutyDate: nil},
	}

	result := FindCurrentDuty(duties, today)
	if result == nil {
		t.Fatal("expected duty, got nil")
	}
	if result.DutyID != "alice" {
		t.Errorf("expected alice (first alphabetically), got %s", result.DutyID)
	}
}

func TestFindCurrentDuty_SinglePerson(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)

	duties := []dao.Duty{
		{ID: 1, DutyID: "alice", LastDutyDate: &yesterday},
	}

	result := FindCurrentDuty(duties, today)
	if result == nil {
		t.Fatal("expected duty, got nil")
	}
	if result.DutyID != "alice" {
		t.Errorf("expected alice (only person), got %s", result.DutyID)
	}
}

func TestFindCurrentDuty_UnsortedInput(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)

	// Input is not sorted alphabetically
	duties := []dao.Duty{
		{ID: 3, DutyID: "charlie", LastDutyDate: nil},
		{ID: 1, DutyID: "alice", LastDutyDate: &yesterday},
		{ID: 2, DutyID: "bob", LastDutyDate: nil},
	}

	result := FindCurrentDuty(duties, today)
	if result == nil {
		t.Fatal("expected duty, got nil")
	}
	if result.DutyID != "bob" {
		t.Errorf("expected bob (next after alice alphabetically), got %s", result.DutyID)
	}
}

func TestFindCurrentDuty_MultipleOldDates(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)
	twoDaysAgo := today.AddDate(0, 0, -2)
	threeDaysAgo := today.AddDate(0, 0, -3)

	duties := []dao.Duty{
		{ID: 1, DutyID: "alice", LastDutyDate: &threeDaysAgo},
		{ID: 2, DutyID: "bob", LastDutyDate: &twoDaysAgo},
		{ID: 3, DutyID: "charlie", LastDutyDate: nil},
	}

	result := FindCurrentDuty(duties, today)
	if result == nil {
		t.Fatal("expected duty, got nil")
	}
	// bob has the most recent date, so next is charlie
	if result.DutyID != "charlie" {
		t.Errorf("expected charlie (next after bob who had most recent duty), got %s", result.DutyID)
	}
}

func TestFindCurrentDuty_SameDayDifferentTime(t *testing.T) {
	today := time.Date(2026, 1, 8, 15, 30, 0, 0, time.UTC)
	todayMorning := time.Date(2026, 1, 8, 9, 0, 0, 0, time.UTC)

	duties := []dao.Duty{
		{ID: 1, DutyID: "alice", LastDutyDate: &todayMorning},
		{ID: 2, DutyID: "bob", LastDutyDate: nil},
	}

	result := FindCurrentDuty(duties, today)
	if result == nil {
		t.Fatal("expected duty, got nil")
	}
	// Should recognize same day even with different times
	if result.DutyID != "alice" {
		t.Errorf("expected alice (same day), got %s", result.DutyID)
	}
}
