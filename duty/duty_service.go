package duty

import (
	"sort"
	"time"

	"watch_bot/dao"
)

// Service handles duty-related business logic
type Service struct {
	connectionStr string
}

// NewService creates a new duty service
func NewService(connectionStr string) *Service {
	return &Service{
		connectionStr: connectionStr,
	}
}

// GetCurrentDuty returns the current duty person and updates the database if needed
// Algorithm:
// 1. Find record where last_duty_date = today -> return it
// 2. If not found, find record with max last_duty_date and get next by duty_id alphabetically
// 3. Update the found record with today's date
func (s *Service) GetCurrentDuty() (*dao.Duty, error) {
	duties, err := dao.GetAllDuties(s.connectionStr)
	if err != nil {
		return nil, err
	}

	currentDate := time.Now().Truncate(24 * time.Hour)
	duty := FindCurrentDuty(duties, currentDate)
	if duty == nil {
		return nil, nil
	}

	// Check if we need to update the database
	if duty.LastDutyDate == nil || !isSameDay(*duty.LastDutyDate, currentDate) {
		err = dao.UpdateDutyDate(s.connectionStr, duty.ID, currentDate)
		if err != nil {
			return nil, err
		}
		duty.LastDutyDate = &currentDate
	}

	return duty, nil
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

func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
