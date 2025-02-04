package watch

import (
	"testing"
	"time"
	"watch_bot/bots"
)

type MockURLChecker struct {
	responses []bool
	index     int
}

func (m *MockURLChecker) IsUrlOk(url string) bool {
	response := m.responses[m.index]
	m.index++
	return response
}

func TestDog(t *testing.T) {
	tests := []struct {
		name               string
		server             Server
		livenessMessages   []string
		unhealthyThreshold int
		expectedMessages   []string
		urlResponses       []bool
	}{
		{
			name: "Server becomes unhealthy",
			server: Server{
				Name: "TestServer",
				URL:  "http://example.com",
			},
			livenessMessages:   []string{"TestServer", "TestServer"},
			unhealthyThreshold: 1,
			expectedMessages:   []string{"TestServer is not OK"},
			urlResponses:       []bool{false, false},
		},
		{
			name: "Server remains healthy",
			server: Server{
				Name: "TestServer",
				URL:  "http://example.com",
			},
			livenessMessages:   []string{"TestServer", "TestServer"},
			unhealthyThreshold: 1,
			expectedMessages:   []string{},
			urlResponses:       []bool{true, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messagesChannel := make(chan bots.Message, len(tt.expectedMessages))
			livenessChannel := make(chan string, len(tt.livenessMessages))
			checker := &MockURLChecker{responses: tt.urlResponses}

			go Dog(tt.server, messagesChannel, "12345", livenessChannel, tt.unhealthyThreshold, checker)

			for _, msg := range tt.livenessMessages {
				livenessChannel <- msg
			}

			close(livenessChannel)

			var receivedMessages []string
			for i := 0; i < len(tt.expectedMessages); i++ {
				select {
				case msg := <-messagesChannel:
					receivedMessages = append(receivedMessages, msg.Text)
				case <-time.After(5 * time.Second):
					t.Fatalf("expected message but got none")
				}
			}

			for i, expected := range tt.expectedMessages {
				if receivedMessages[i] != expected {
					t.Errorf("expected %v, got %v", expected, receivedMessages[i])
				}
			}
		})
	}
}
