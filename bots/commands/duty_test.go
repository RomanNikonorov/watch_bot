package commands

import (
	"strings"
	"testing"
	"watch_bot/bots"
	"watch_bot/duty"
)

// mockDutyService is a test double for dutyServicer
type mockDutyService struct {
	result *duty.DutyResult
	err    error
}

func (m *mockDutyService) GetCurrentDuty() (*duty.DutyResult, error) {
	return m.result, m.err
}

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

func TestDutyCommand_Execute_SupportChatMessageContainsAtPrefix(t *testing.T) {
	messagesChan := make(chan bots.Message, 10)
	cmd := &DutyCommand{
		dutyService:   &mockDutyService{result: &duty.DutyResult{DutyID: "johndoe"}},
		messagesChan:  messagesChan,
		supportChatId: "support-123",
	}

	_, err := cmd.Execute(bots.Command{
		Name:   "duty",
		ChatId: "caller-456",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	close(messagesChan)

	var supportMsg *bots.Message
	for msg := range messagesChan {
		if msg.ChatId == "support-123" {
			msg := msg
			supportMsg = &msg
		}
	}
	if supportMsg == nil {
		t.Fatal("expected a message to be sent to the support chat")
	}
	// Check that message contains mention in @[userId] format
	expectedMention := "@[johndoe]"
	if !strings.Contains(supportMsg.Text, expectedMention) {
		t.Errorf("expected support chat message to contain %s, got: %s", expectedMention, supportMsg.Text)
	}
	// Check that ParseMode is set to HTML
	if supportMsg.ParseMode != "HTML" {
		t.Errorf("expected ParseMode to be HTML, got: %s", supportMsg.ParseMode)
	}
}
