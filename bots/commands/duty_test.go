package commands

import (
	"testing"
	"watch_bot/bots"
)

func TestDutyCommand_Description(t *testing.T) {
	cmd := NewDutyCommand(DutyCommandConfig{
		ConnectionStr: "",
		MessagesChan:  nil,
		SupportChatId: "",
	})
	desc := cmd.Description()
	if desc == "" {
		t.Error("expected non-empty description")
	}
}

func TestDutyCommand_Execute_NoConnection(t *testing.T) {
	cmd := NewDutyCommand(DutyCommandConfig{
		ConnectionStr: "invalid_connection_string",
		MessagesChan:  nil,
		SupportChatId: "",
	})
	_, err := cmd.Execute(bots.Command{
		Name:   "duty",
		ChatId: "123",
		Params: map[string]string{},
	})
	// Should return error due to invalid connection
	if err == nil {
		t.Error("expected error for invalid connection string")
	}
}
