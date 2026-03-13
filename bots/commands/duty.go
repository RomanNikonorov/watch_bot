package commands

import (
	"fmt"
	"log"
	"watch_bot/bots"
	"watch_bot/duty"
)

// DutyCommandConfig contains configuration for the duty command
type DutyCommandConfig struct {
	ConnectionStr string
	MessagesChan  chan bots.Message
	SupportChatId string
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
}

// NewDutyCommand creates a new DutyCommand
func NewDutyCommand(config DutyCommandConfig) *DutyCommand {
	return &DutyCommand{
		dutyService:   duty.NewService(config.ConnectionStr),
		messagesChan:  config.MessagesChan,
		supportChatId: config.SupportChatId,
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

		// Send notification to support chat about who is on duty
		if d.supportChatId != "" {
			// Use MarkdownV2 format for mentions: [userId](mention://userId)
			// Note: VK Teams mention format - test if this works, alternatives include:
			// - @[userId](mention://userId) with @ prefix
			// - HTML format: <a href="mention://userId">userId</a> with ParseModeHTML
			notificationText := fmt.Sprintf("⚠️ Duty person called!\n\nOn duty today: [%s](mention://%s)", result.DutyID, result.DutyID)
			select {
			case d.messagesChan <- bots.Message{
				ChatId:    d.supportChatId,
				Text:      notificationText,
				ParseMode: "MarkdownV2",
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
