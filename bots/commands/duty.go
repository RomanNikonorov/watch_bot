package commands

import (
	"fmt"
	"log"
	"watch_bot/bots"
	"watch_bot/duty"

	botgolang "github.com/mail-ru-im/bot-golang"
)

// DutyCommandConfig contains configuration for the duty command
type DutyCommandConfig struct {
	ConnectionStr string
	MessagesChan  chan bots.Message
	SupportChatId string
	IsWorkingNow  func() bool
}

// dutyServicer is the interface for retrieving current duty information
type dutyServicer interface {
	GetCurrentDuty() (*duty.DutyResult, error)
}

// DutyCommand handles the \duty command
type DutyCommand struct {
	dutyService   dutyServicer
	messagesChan  chan bots.Message
	supportChatId string
	isWorkingNow  func() bool
}

// NewDutyCommand creates a new DutyCommand
func NewDutyCommand(config DutyCommandConfig) *DutyCommand {
	return &DutyCommand{
		dutyService:   duty.NewService(config.ConnectionStr),
		messagesChan:  config.MessagesChan,
		supportChatId: config.SupportChatId,
		isWorkingNow:  config.IsWorkingNow,
	}
}

// Execute handles the duty command
func (d *DutyCommand) Execute(cmd bots.Command) (string, error) {
	if d.isWorkingNow != nil && !d.isWorkingNow() {
		return "Duty can only be called during working hours", nil
	}

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

		// Send notification to support chat about who is on duty (only on first assignment of the day)
		if d.supportChatId != "" && result.IsNewAssignment {
			// Use @[userId] format for mentions in VK Teams with HTML ParseMode
			notificationText := fmt.Sprintf("⚠️ Duty person called!\n\nOn duty today: @[%s]", result.DutyID)
			select {
			case d.messagesChan <- bots.Message{
				ChatId:    d.supportChatId,
				Text:      notificationText,
				ParseMode: string(botgolang.ParseModeHTML),
			}:
				// Message sent successfully
			default:
				log.Printf("Warning: failed to send notification to support chat %s: channel buffer full", d.supportChatId)
			}
		}
	}

	return "The development team is rushing to help!", nil
}

// Description returns command description
func (d *DutyCommand) Description() string {
	return "show current duty person"
}
