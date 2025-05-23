package working_calendar

import (
	"os"
	"reflect"
	"testing"
	"time"
)

func TestFillWorkingTime(t *testing.T) {
	location, _ := time.LoadLocation("Local")
	startTime, _ := time.ParseInLocation("15:04", "09:00", location)
	endTime, _ := time.ParseInLocation("15:04", "18:00", location)

	tests := []struct {
		name    string
		envVars map[string]string
		want    WorkingTime
	}{
		{
			name: "valid working time",
			envVars: map[string]string{
				"START_TIME": "09:00",
				"END_TIME":   "18:00",
				"DAYS_OFF":   "Saturday,Sunday",
			},
			want: WorkingTime{
				StartTime:      startTime,
				EndTime:        endTime,
				DaysOff:        []time.Weekday{time.Saturday, time.Sunday},
				hasWorkingTime: true,
			},
		},
		{
			name: "invalid time format",
			envVars: map[string]string{
				"START_TIME": "25:00",
				"END_TIME":   "18:00",
				"DAYS_OFF":   "Saturday,Sunday",
			},
			want: WorkingTime{
				hasWorkingTime: false,
			},
		},
		{
			name: "empty environment variables",
			envVars: map[string]string{
				"START_TIME": "",
				"END_TIME":   "",
				"DAYS_OFF":   "",
			},
			want: WorkingTime{
				hasWorkingTime: false,
			},
		},
		{
			name: "multiple days off with spaces",
			envVars: map[string]string{
				"START_TIME": "09:00",
				"END_TIME":   "18:00",
				"DAYS_OFF":   "Saturday, Sunday, Monday",
			},
			want: WorkingTime{
				StartTime:      startTime,
				EndTime:        endTime,
				DaysOff:        []time.Weekday{time.Saturday, time.Sunday, time.Monday},
				hasWorkingTime: true,
			},
		},
		{
			name: "days off not set",
			envVars: map[string]string{
				"START_TIME": "09:00",
				"END_TIME":   "18:00",
				"DAYS_OFF":   "",
			},
			want: WorkingTime{
				StartTime:      startTime,
				EndTime:        endTime,
				DaysOff:        nil,
				hasWorkingTime: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Cleanup
			defer func() {
				for k := range tt.envVars {
					os.Unsetenv(k)
				}
			}()

			got := FillWorkingTime()

			// Check hasWorkingTime
			if got.hasWorkingTime != tt.want.hasWorkingTime {
				t.Errorf("FillWorkingTime().hasWorkingTime = %v, want %v",
					got.hasWorkingTime, tt.want.hasWorkingTime)
			}

			if got.hasWorkingTime {
				// Check times
				if !got.StartTime.Equal(tt.want.StartTime) {
					t.Errorf("FillWorkingTime().StartTime = %v, want %v",
						got.StartTime, tt.want.StartTime)
				}
				if !got.EndTime.Equal(tt.want.EndTime) {
					t.Errorf("FillWorkingTime().EndTime = %v, want %v",
						got.EndTime, tt.want.EndTime)
				}
				// Check days off
				if !reflect.DeepEqual(got.DaysOff, tt.want.DaysOff) {
					t.Errorf("FillWorkingTime().DaysOff = %v, want %v",
						got.DaysOff, tt.want.DaysOff)
				}
			}
		})
	}
}

func TestIsWorkingTime(t *testing.T) {
	location, _ := time.LoadLocation("Local")
	// Use fixed times for test predictability
	startTime, _ := time.ParseInLocation("15:04", "09:00", location)
	endTime, _ := time.ParseInLocation("15:04", "18:00", location)

	tests := []struct {
		name        string
		workingTime WorkingTime
		currentTime time.Time
		want        bool
	}{
		{
			name: "during working hours on working day",
			workingTime: WorkingTime{
				StartTime:      startTime,
				EndTime:        endTime,
				DaysOff:        []time.Weekday{time.Saturday, time.Sunday},
				hasWorkingTime: true,
			},
			currentTime: time.Date(2024, 3, 20, 14, 30, 0, 0, location), // Wednesday
			want:        true,
		},
		{
			name: "before working hours on working day",
			workingTime: WorkingTime{
				StartTime:      startTime,
				EndTime:        endTime,
				DaysOff:        []time.Weekday{time.Saturday, time.Sunday},
				hasWorkingTime: true,
			},
			currentTime: time.Date(2024, 3, 20, 8, 59, 0, 0, location),
			want:        false,
		},

		{
			name: "days off not set",
			workingTime: WorkingTime{
				StartTime:      startTime,
				EndTime:        endTime,
				DaysOff:        nil,
				hasWorkingTime: true,
			},
			currentTime: time.Date(2024, 3, 20, 8, 59, 0, 0, location),
			want:        false,
		},
		{
			name: "during working hours on day off",
			workingTime: WorkingTime{
				StartTime:      startTime,
				EndTime:        endTime,
				DaysOff:        []time.Weekday{time.Saturday, time.Sunday},
				hasWorkingTime: true,
			},
			currentTime: time.Date(2024, 3, 23, 14, 30, 0, 0, location), // Saturday
			want:        false,
		},
		{
			name: "working time not set",
			workingTime: WorkingTime{
				hasWorkingTime: false,
			},
			currentTime: time.Date(2024, 3, 20, 14, 30, 0, 0, location),
			want:        true,
		},
		{
			name: "multiple days off",
			workingTime: WorkingTime{
				StartTime:      startTime,
				EndTime:        endTime,
				DaysOff:        []time.Weekday{time.Saturday, time.Sunday, time.Friday},
				hasWorkingTime: true,
			},
			currentTime: time.Date(2024, 3, 22, 14, 30, 0, 0, location), // Friday
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsWorkingTime(tt.workingTime, tt.currentTime, []time.Time{})
			if got != tt.want {
				t.Errorf("IsWorkingTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsUnusualDay(t *testing.T) {
	location, _ := time.LoadLocation("Local")

	tests := []struct {
		name        string
		currentTime time.Time
		unusualDays []time.Time
		want        bool
	}{
		{
			name:        "current day is unusual",
			currentTime: time.Date(2024, 5, 1, 14, 30, 0, 0, location),
			unusualDays: []time.Time{
				time.Date(2024, 4, 30, 0, 0, 0, 0, location),
				time.Date(2024, 5, 1, 0, 0, 0, 0, location), // May 1st is unusual
				time.Date(2024, 5, 2, 0, 0, 0, 0, location),
			},
			want: true,
		},
		{
			name:        "current day is not unusual",
			currentTime: time.Date(2024, 5, 3, 10, 0, 0, 0, location),
			unusualDays: []time.Time{
				time.Date(2024, 5, 1, 0, 0, 0, 0, location),
				time.Date(2024, 5, 2, 0, 0, 0, 0, location),
			},
			want: false,
		},
		{
			name:        "empty unusual days list",
			currentTime: time.Date(2024, 5, 1, 12, 0, 0, 0, location),
			unusualDays: []time.Time{},
			want:        false,
		},
		{
			name:        "current time with non-zero hour/minute ignores time part",
			currentTime: time.Date(2024, 5, 1, 14, 30, 45, 0, location),
			unusualDays: []time.Time{
				time.Date(2024, 5, 1, 0, 0, 0, 0, location),
			},
			want: true,
		},
		{
			name:        "unusual day with non-zero time part",
			currentTime: time.Date(2024, 5, 1, 0, 0, 0, 0, location),
			unusualDays: []time.Time{
				time.Date(2024, 5, 1, 10, 30, 0, 0, location),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUnusualDay(tt.currentTime, tt.unusualDays)
			if got != tt.want {
				t.Errorf("isUnusualDay() = %v, want %v", got, tt.want)
			}
		})
	}
}
