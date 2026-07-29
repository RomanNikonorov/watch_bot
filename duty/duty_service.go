package duty

import (
	"context"
	"sort"
	"time"

	"watch_bot/dao"
	"watch_bot/lib"
)

// Service handles duty-related business logic
type Service struct {
	store *dao.Store
	now   func() time.Time
}

// NewService creates a new duty service
func NewService(store *dao.Store) *Service {
	return &Service{
		store: store,
		now:   time.Now,
	}
}

// DutyResult contains information about the current duty person
type DutyResult struct {
	DutyID          string // ID of the duty chat
	IsNewAssignment bool   // True if duty was just assigned today (first call of the day)
}

// GetCurrentDuty returns the current duty person and updates the database if needed
// Algorithm:
// 1. Find record where last_duty_date = today -> return it
// 2. If not found, find record with max last_duty_date and get next by duty_id alphabetically
// 3. Update the found record with today's date
func (s *Service) GetCurrentDuty(ctx context.Context) (*DutyResult, error) {
	currentDate := lib.StartOfDay(s.now())
	var result *DutyResult

	err := s.store.WithDutyRotation(ctx, func(duties []dao.Duty, rotation *dao.DutyRotation) error {
		duty := FindCurrentDuty(duties, currentDate)
		if duty == nil {
			return nil
		}

		isNewAssignment := duty.LastDutyDate == nil || !isSameDay(*duty.LastDutyDate, currentDate)
		if isNewAssignment {
			if err := rotation.SetDutyDate(ctx, duty.ID, currentDate); err != nil {
				return err
			}
		}

		result = &DutyResult{
			DutyID:          duty.DutyID,
			IsNewAssignment: isNewAssignment,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetNextDuty forcefully moves today's duty to the next person in alphabetical rotation.
func (s *Service) GetNextDuty(ctx context.Context) (*DutyResult, error) {
	currentDate := lib.StartOfDay(s.now())
	var result *DutyResult

	err := s.store.WithDutyRotation(ctx, func(duties []dao.Duty, rotation *dao.DutyRotation) error {
		duty := FindNextDuty(duties, currentDate)
		if duty == nil {
			return nil
		}
		if err := rotation.ReassignDutyDate(ctx, duty.ID, currentDate); err != nil {
			return err
		}

		result = &DutyResult{
			DutyID:          duty.DutyID,
			IsNewAssignment: true,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// FindCurrentDuty finds the current duty person from a list of duties
// This is a pure function for easy testing
func FindCurrentDuty(duties []dao.Duty, currentDate time.Time) *dao.Duty {
	if len(duties) == 0 {
		return nil
	}

	// Ensure duties are sorted by duty_id
	sort.Slice(duties, func(i, j int) bool {
		return duties[i].DutyID < duties[j].DutyID
	})

	// Step 1: Find duty for today
	for i := range duties {
		if duties[i].LastDutyDate != nil && isSameDay(*duties[i].LastDutyDate, currentDate) {
			return &duties[i]
		}
	}

	// Step 2: Find the last duty person (with max last_duty_date)
	var lastDutyIndex = -1
	var maxDate time.Time
	for i := range duties {
		if duties[i].LastDutyDate != nil {
			if lastDutyIndex == -1 || duties[i].LastDutyDate.After(maxDate) {
				maxDate = *duties[i].LastDutyDate
				lastDutyIndex = i
			}
		}
	}

	// Step 3: Get next person alphabetically, or first if wrap around
	if lastDutyIndex == -1 {
		// No one has been on duty yet, return first
		return &duties[0]
	}

	nextIndex := (lastDutyIndex + 1) % len(duties)
	return &duties[nextIndex]
}

// FindNextDuty finds the next duty person after today's assigned duty.
func FindNextDuty(duties []dao.Duty, currentDate time.Time) *dao.Duty {
	if len(duties) < 2 {
		return nil
	}

	sort.Slice(duties, func(i, j int) bool {
		return duties[i].DutyID < duties[j].DutyID
	})

	currentDutyIndex := -1
	for i := range duties {
		if duties[i].LastDutyDate != nil && isSameDay(*duties[i].LastDutyDate, currentDate) {
			currentDutyIndex = i
			break
		}
	}

	if currentDutyIndex == -1 {
		return nil
	}

	nextIndex := (currentDutyIndex + 1) % len(duties)
	return &duties[nextIndex]
}

func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
