package working_calendar

import (
	"log"
	"os"
	"strings"
	"time"
)

type WorkingTime struct {
	StartTime      time.Time
	EndTime        time.Time
	DaysOff        []time.Weekday
	hasWorkingTime bool
}

func contains(weekdays []time.Weekday, day time.Weekday) bool {
	for _, d := range weekdays {
		if d == day {
			return true
		}
	}
	return false
}

func IsWorkingTime(workingTime WorkingTime, currentTime time.Time, unusualDays []time.Time) bool {

	if !workingTime.hasWorkingTime {
		return true
	}

	currentWeekday := currentTime.Weekday()

	isUnusualDay := isUnusualDay(currentTime, unusualDays)

	if contains(workingTime.DaysOff, currentWeekday) {
		return isUnusualDay
	}

	currentTimeTimeOnly := time.Date(0, 1, 1, currentTime.Hour(), currentTime.Minute(), currentTime.Second(), currentTime.Nanosecond(), currentTime.Location())
	if currentTimeTimeOnly.Before(workingTime.StartTime) || currentTimeTimeOnly.After(workingTime.EndTime) {
		return false
	}

	return !isUnusualDay

}

// isUnusualDay checks if the current date (ignoring time) matches any date in the unusual days list
func isUnusualDay(currentTime time.Time, unusualDays []time.Time) bool {
	// If no unusual days defined, return false
	if len(unusualDays) == 0 {
		return false
	}

	location, errLocation := time.LoadLocation("Local")

	if errLocation != nil {
		log.Printf("Error loading location: %v", errLocation)
		location = time.UTC // fallback to UTC if location cannot be loaded
	}

	// remove time component from current time, keeping only the date part
	currentDate := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, location)

	for _, unusualDay := range unusualDays {
		// remove time component from unusual day, keeping only the date part
		unusualDate := time.Date(unusualDay.Year(), unusualDay.Month(), unusualDay.Day(), 0, 0, 0, 0, location)

		if currentDate.Equal(unusualDate) {
			return true
		}
	}

	return false
}

func FillWorkingTime() WorkingTime {
	startTimeStr := os.Getenv("START_TIME")
	endTimeStr := os.Getenv("END_TIME")
	daysOffStr := os.Getenv("DAYS_OFF")

	location, errLocation := time.LoadLocation("Local")
	if errLocation != nil {
		log.Printf("Will not check working time. Error parsing location: %v", errLocation)
		return WorkingTime{
			hasWorkingTime: false,
		}
	}

	startTime, errStartTime := time.ParseInLocation("15:04", startTimeStr, location)
	endTime, errEndTime := time.ParseInLocation("15:04", endTimeStr, location)
	if errStartTime != nil {
		log.Printf("Will not check working time. Error parsing start time: %v", errStartTime)
		return WorkingTime{
			hasWorkingTime: false,
		}
	}
	if errEndTime != nil {
		log.Printf("Will not check working time. Error parsing end time: %v", errEndTime)
		return WorkingTime{
			hasWorkingTime: false,
		}
	}

	var daysOff []time.Weekday
	if daysOffStr != "" {
		daysOffStrSlice := strings.Split(daysOffStr, ",")
		daysOff = make([]time.Weekday, len(daysOffStrSlice))
		for i, day := range daysOffStrSlice {
			switch strings.TrimSpace(day) {
			case "Sunday":
				daysOff[i] = time.Sunday
			case "Monday":
				daysOff[i] = time.Monday
			case "Tuesday":
				daysOff[i] = time.Tuesday
			case "Wednesday":
				daysOff[i] = time.Wednesday
			case "Thursday":
				daysOff[i] = time.Thursday
			case "Friday":
				daysOff[i] = time.Friday
			case "Saturday":
				daysOff[i] = time.Saturday
			default:
				log.Printf("Unknown day off: %v", day)
			}
		}
	}

	log.Printf("Working time: %v - %v, days off: %v", startTime.Format("15:04"), endTime.Format("15:04"), daysOff)
	return WorkingTime{
		StartTime:      startTime,
		EndTime:        endTime,
		DaysOff:        daysOff,
		hasWorkingTime: true,
	}
}
