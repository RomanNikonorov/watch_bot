package bots

import (
	"testing"
)

func TestParseCommand_ValidCommand(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		chatId   string
		expected *Command
	}{
		{
			name:   "simple command",
			text:   "\\duty",
			chatId: "123",
			expected: &Command{
				Name:   "duty",
				ChatId: "123",
				Params: map[string]string{},
			},
		},
		{
			name:   "command with params",
			text:   "\\duty param1 param2",
			chatId: "456",
			expected: &Command{
				Name:   "duty",
				ChatId: "456",
				Params: map[string]string{"0": "param1", "1": "param2"},
			},
		},
		{
			name:   "command with extra spaces",
			text:   "  \\duty   param1  ",
			chatId: "789",
			expected: &Command{
				Name:   "duty",
				ChatId: "789",
				Params: map[string]string{"0": "param1"},
			},
		},
		{
			name:   "uppercase command normalized to lowercase",
			text:   "\\DUTY",
			chatId: "123",
			expected: &Command{
				Name:   "duty",
				ChatId: "123",
				Params: map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseCommand(tt.text, tt.chatId)
			if result == nil {
				t.Fatalf("expected command, got nil")
			}
			if result.Name != tt.expected.Name {
				t.Errorf("expected name %s, got %s", tt.expected.Name, result.Name)
			}
			if result.ChatId != tt.expected.ChatId {
				t.Errorf("expected chatId %s, got %s", tt.expected.ChatId, result.ChatId)
			}
			if len(result.Params) != len(tt.expected.Params) {
				t.Errorf("expected %d params, got %d", len(tt.expected.Params), len(result.Params))
			}
			for k, v := range tt.expected.Params {
				if result.Params[k] != v {
					t.Errorf("expected param %s=%s, got %s", k, v, result.Params[k])
				}
			}
		})
	}
}

func TestParseCommand_NotACommand(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		chatId string
	}{
		{
			name:   "regular message",
			text:   "hello world",
			chatId: "123",
		},
		{
			name:   "empty string",
			text:   "",
			chatId: "123",
		},
		{
			name:   "only backslash",
			text:   "\\",
			chatId: "123",
		},
		{
			name:   "backslash with spaces only",
			text:   "\\   ",
			chatId: "123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseCommand(tt.text, tt.chatId)
			if result != nil {
				t.Errorf("expected nil, got command: %+v", result)
			}
		})
	}
}

// Mock handler for testing
type mockHandler struct {
	response string
	err      error
}

func (m *mockHandler) Execute(cmd Command) (string, error) {
	return m.response, m.err
}

func (m *mockHandler) Description() string {
	return "mock command"
}

func TestCommandRouter_Handle(t *testing.T) {
	router := NewCommandRouter()
	router.Register("test", &mockHandler{response: "test response", err: nil})

	t.Run("known command", func(t *testing.T) {
		cmd := Command{Name: "test", ChatId: "123", Params: map[string]string{}}
		response, err := router.Handle(cmd)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if response != "test response" {
			t.Errorf("expected 'test response', got '%s'", response)
		}
	})

	t.Run("unknown command", func(t *testing.T) {
		cmd := Command{Name: "unknown", ChatId: "123", Params: map[string]string{}}
		response, err := router.Handle(cmd)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if response == "" {
			t.Error("expected non-empty response for unknown command")
		}
	})
}

func TestCommandRouter_GetRegisteredCommands(t *testing.T) {
	router := NewCommandRouter()
	router.Register("cmd1", &mockHandler{response: "1"})
	router.Register("cmd2", &mockHandler{response: "2"})

	commands := router.GetRegisteredCommands()
	if len(commands) != 2 {
		t.Errorf("expected 2 commands, got %d", len(commands))
	}
	if _, exists := commands["cmd1"]; !exists {
		t.Error("expected cmd1 to be registered")
	}
	if _, exists := commands["cmd2"]; !exists {
		t.Error("expected cmd2 to be registered")
	}
}
