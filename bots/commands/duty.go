package commands

import (
	"fmt"
	"watch_bot/bots"
	"watch_bot/duty"
)

// DutyCommand handles the \duty command
type DutyCommand struct {
	dutyService  *duty.Service
	messagesChan chan bots.Message
}

// NewDutyCommand creates a new DutyCommand
func NewDutyCommand(connectionStr string, messagesChan chan bots.Message) *DutyCommand {
	return &DutyCommand{
		dutyService:  duty.NewService(connectionStr),
		messagesChan: messagesChan,
	}
}

// Execute handles the duty command
func (d *DutyCommand) Execute(cmd bots.Command) (string, error) {
	result, err := d.dutyService.GetCurrentDuty()
	if err != nil {
		return "", fmt.Errorf("failed to get current duty: %w", err)
	}
	if result == nil {
		return "No duty assigned for today", nil
	}

	// Send notification to the duty person via channel
	if d.messagesChan != nil {
		d.messagesChan <- bots.Message{
			ChatId: result.DutyID,
			Text:   "You are on duty today!",
		}
	}

	return "The development team is rushing to help!", nil
}

// Description returns command description
func (d *DutyCommand) Description() string {
	return "show current duty person"
}
