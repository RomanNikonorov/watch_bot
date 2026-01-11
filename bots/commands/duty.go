package commands

import (
	"fmt"
	"log"
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

	// Send notification to the duty person via channel (non-blocking)
	if d.messagesChan != nil {
		select {
		case d.messagesChan <- bots.Message{
			ChatId: result.DutyID,
			Text:   "You are on duty today!",
		}:
			// Message sent successfully
		default:
			log.Printf("Warning: failed to send notification to duty person %s: channel buffer full", result.DutyID)
		}
	}

	return "The development team is rushing to help!", nil
}

// Description returns command description
func (d *DutyCommand) Description() string {
	return "show current duty person"
}
