package commands

import (
	"context"
	"fmt"
	"watch_bot/bots"
	"watch_bot/dao"
	"watch_bot/duty"
)

// DutyCommandConfig contains configuration for the duty command
type DutyCommandConfig struct {
	Store         *dao.Store
	MessagesChan  chan bots.Message
	SupportChatId string
	BotType       string
	IsWorkingNow  func() bool
}

// dutyServicer is the interface for retrieving current duty information
type dutyServicer interface {
	GetCurrentDuty(context.Context) (*duty.DutyResult, error)
}

// DutyCommand handles the \duty command
type DutyCommand struct {
	dutyService   dutyServicer
	messagesChan  chan bots.Message
	supportChatId string
	botType       string
	isWorkingNow  func() bool
}

// NewDutyCommand creates a new DutyCommand
func NewDutyCommand(config DutyCommandConfig) *DutyCommand {
	return &DutyCommand{
		dutyService:   duty.NewService(config.Store),
		messagesChan:  config.MessagesChan,
		supportChatId: config.SupportChatId,
		botType:       config.BotType,
		isWorkingNow:  config.IsWorkingNow,
	}
}

// Execute handles the duty command
func (d *DutyCommand) Execute(ctx context.Context, cmd bots.Command) (string, error) {
	if d.isWorkingNow != nil && !d.isWorkingNow() {
		return "Duty can only be called during working hours", nil
	}

	result, err := d.dutyService.GetCurrentDuty(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get current duty: %w", err)
	}
	if result == nil {
		return "No duty assigned for today", nil
	}

	if d.messagesChan != nil {
		if err := bots.EnqueueMessage(ctx, d.messagesChan, bots.Message{
			ChatId: result.DutyID,
			Text:   "You are on duty today!",
		}); err != nil {
			return "", fmt.Errorf("failed to enqueue duty notification: %w", err)
		}

		// Send notification to support chat about who is on duty (only on first assignment of the day)
		if d.supportChatId != "" && result.IsNewAssignment {
			mention, parseMode := bots.FormatMention(d.botType, result.DutyID)
			notificationText := fmt.Sprintf("⚠️ Duty person called!\n\nOn duty today: %s", mention)
			if err := bots.EnqueueMessage(ctx, d.messagesChan, bots.Message{
				ChatId:    d.supportChatId,
				Text:      notificationText,
				ParseMode: parseMode,
			}); err != nil {
				return "", fmt.Errorf("failed to enqueue support notification: %w", err)
			}
		}
	}

	return "The development team is rushing to help!", nil
}

// Description returns command description
func (d *DutyCommand) Description() string {
	return "show current duty person"
}
