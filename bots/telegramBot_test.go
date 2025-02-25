package bots

import (
	"errors"
	"testing"
	"time"
)

func TestTgSendWithRetry(t *testing.T) {
	tests := []struct {
		name       string
		failCount  int
		retryCount int
		pause      int
		wantTries  int
	}{
		{
			name:       "success on first attempt",
			failCount:  0,
			retryCount: 3,
			pause:      1,
			wantTries:  1,
		},
		{
			name:       "success after one retry",
			failCount:  1,
			retryCount: 3,
			pause:      1,
			wantTries:  2,
		},
		{
			name:       "all attempts fail",
			failCount:  3,
			retryCount: 3,
			pause:      1,
			wantTries:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sendCount := 0
			sendFunc := func() error {
				sendCount++
				if sendCount <= tt.failCount {
					return errors.New("send failed")
				}
				return nil
			}

			start := time.Now()
			tgSendWithRetry(sendFunc, tt.retryCount, tt.pause)

			if sendCount != tt.wantTries {
				t.Errorf("got %d attempts, want %d", sendCount, tt.wantTries)
			}

			duration := time.Since(start)
			expectedDuration := time.Duration(tt.failCount) * time.Duration(tt.pause) * time.Second
			if duration < expectedDuration {
				t.Errorf("duration %v shorter than expected %v", duration, expectedDuration)
			}
		})
	}
}
