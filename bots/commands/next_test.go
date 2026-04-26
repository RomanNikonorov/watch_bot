package commands

import (
	"strings"
	"testing"
	"watch_bot/bots"
	"watch_bot/duty"
)

type mockNextDutyService struct {
	result *duty.DutyResult
	err    error
}

func (m *mockNextDutyService) GetNextDuty() (*duty.DutyResult, error) {
	return m.result, m.err
}

func TestNextCommand_Execute_SendsDutyAndSupportMessages(t *testing.T) {
	messagesChan := make(chan bots.Message, 10)
	cmd := &NextCommand{
		dutyService:        &mockNextDutyService{result: &duty.DutyResult{DutyID: "janedoe", IsNewAssignment: true}},
		messagesChan:       messagesChan,
		supportChatId:      "support-123",
		allowedNextUserIds: newAllowedUserIds([]string{"user-1"}),
		isWorkingNow: func() bool {
			return true
		},
	}

	response, err := cmd.Execute(bots.Command{
		Name:   "next",
		ChatId: "support-123",
		UserId: "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response != "Next duty person has been selected" {
		t.Fatalf("unexpected response: %q", response)
	}

	close(messagesChan)

	var directMsg *bots.Message
	var supportMsg *bots.Message
	for msg := range messagesChan {
		msg := msg
		if msg.ChatId == "janedoe" {
			directMsg = &msg
		}
		if msg.ChatId == "support-123" {
			supportMsg = &msg
		}
	}

	if directMsg == nil {
		t.Fatal("expected a direct message to be sent to the new duty person")
	}
	if supportMsg == nil {
		t.Fatal("expected a support chat message")
	}
	if !strings.Contains(supportMsg.Text, "@[janedoe]") {
		t.Fatalf("expected support message to mention janedoe, got %q", supportMsg.Text)
	}
	if supportMsg.ParseMode != "HTML" {
		t.Fatalf("expected HTML parse mode, got %q", supportMsg.ParseMode)
	}
}

func TestNextCommand_Execute_NoNextDuty(t *testing.T) {
	cmd := &NextCommand{
		dutyService:        &mockNextDutyService{result: nil},
		allowedNextUserIds: newAllowedUserIds([]string{"user-1"}),
		isWorkingNow: func() bool {
			return true
		},
	}

	response, err := cmd.Execute(bots.Command{Name: "next", ChatId: "support-123", UserId: "user-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response != "No next duty person available for today" {
		t.Fatalf("unexpected response: %q", response)
	}
}

func TestNextCommand_Execute_RejectsOutsideWorkingHours(t *testing.T) {
	messagesChan := make(chan bots.Message, 10)
	cmd := &NextCommand{
		dutyService:        &mockNextDutyService{result: &duty.DutyResult{DutyID: "janedoe", IsNewAssignment: true}},
		messagesChan:       messagesChan,
		allowedNextUserIds: newAllowedUserIds([]string{"user-1"}),
		isWorkingNow: func() bool {
			return false
		},
	}

	response, err := cmd.Execute(bots.Command{Name: "next", ChatId: "support-123", UserId: "user-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response != "Duty can only be changed during working hours" {
		t.Fatalf("unexpected response: %q", response)
	}
	if len(messagesChan) != 0 {
		t.Fatalf("expected no outgoing messages outside working hours, got %d", len(messagesChan))
	}
}

func TestNextCommand_Execute_RejectsUnauthorizedUser(t *testing.T) {
	messagesChan := make(chan bots.Message, 10)
	cmd := &NextCommand{
		dutyService:        &mockNextDutyService{result: &duty.DutyResult{DutyID: "janedoe", IsNewAssignment: true}},
		messagesChan:       messagesChan,
		allowedNextUserIds: newAllowedUserIds([]string{"user-1"}),
		isWorkingNow: func() bool {
			return true
		},
	}

	response, err := cmd.Execute(bots.Command{Name: "next", ChatId: "support-123", UserId: "user-2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response != "You are not allowed to execute this command" {
		t.Fatalf("unexpected response: %q", response)
	}
	if len(messagesChan) != 0 {
		t.Fatalf("expected no outgoing messages for unauthorized user, got %d", len(messagesChan))
	}
}

func TestNextCommand_Execute_RejectsWhenAllowedUsersEmpty(t *testing.T) {
	cmd := &NextCommand{
		dutyService: &mockNextDutyService{result: &duty.DutyResult{DutyID: "janedoe", IsNewAssignment: true}},
		isWorkingNow: func() bool {
			return true
		},
	}

	response, err := cmd.Execute(bots.Command{Name: "next", ChatId: "support-123", UserId: "user-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response != "You are not allowed to execute this command" {
		t.Fatalf("unexpected response: %q", response)
	}
}
