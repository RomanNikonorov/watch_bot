package commands

import (
	"context"
	"fmt"
	"strings"
	"watch_bot/bots"
	"watch_bot/dao"
	"watch_bot/duty"
)

type NextCommandConfig struct {
	Store              *dao.Store
	MessagesChan       chan bots.Message
	SupportChatId      string
	BotType            string
	AllowedNextUserIds []string
	IsWorkingNow       func() bool
}

type nextDutyServicer interface {
	GetNextDuty(context.Context) (*duty.DutyResult, error)
}

type NextCommand struct {
	dutyService        nextDutyServicer
	messagesChan       chan bots.Message
	supportChatId      string
	botType            string
	allowedNextUserIds map[string]struct{}
	isWorkingNow       func() bool
}

func NewNextCommand(config NextCommandConfig) *NextCommand {
	return &NextCommand{
		dutyService:        duty.NewService(config.Store),
		messagesChan:       config.MessagesChan,
		supportChatId:      config.SupportChatId,
		botType:            config.BotType,
		allowedNextUserIds: newAllowedUserIds(config.AllowedNextUserIds),
		isWorkingNow:       config.IsWorkingNow,
	}
}

func (n *NextCommand) Execute(ctx context.Context, cmd bots.Command) (string, error) {
	if !n.isUserAllowed(cmd.UserId) {
		return "You are not allowed to execute this command", nil
	}

	if n.isWorkingNow != nil && !n.isWorkingNow() {
		return "Duty can only be changed during working hours", nil
	}

	result, err := n.dutyService.GetNextDuty(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get next duty: %w", err)
	}
	if result == nil {
		return "No next duty person available for today", nil
	}

	if n.messagesChan != nil {
		if err := bots.EnqueueMessage(ctx, n.messagesChan, bots.Message{
			ChatId: result.DutyID,
			Text:   "You are on duty today!",
		}); err != nil {
			return "", fmt.Errorf("failed to enqueue duty notification: %w", err)
		}

		if n.supportChatId != "" {
			mention, parseMode := bots.FormatMention(n.botType, result.DutyID)
			notificationText := fmt.Sprintf("Duty person changed.\n\nOn duty now: %s", mention)
			if err := bots.EnqueueMessage(ctx, n.messagesChan, bots.Message{
				ChatId:    n.supportChatId,
				Text:      notificationText,
				ParseMode: parseMode,
			}); err != nil {
				return "", fmt.Errorf("failed to enqueue support notification: %w", err)
			}
		}
	}

	return "Next duty person has been selected", nil
}

func (n *NextCommand) Description() string {
	return "assign next duty person"
}

func (n *NextCommand) isUserAllowed(userId string) bool {
	if len(n.allowedNextUserIds) == 0 {
		return false
	}
	_, ok := n.allowedNextUserIds[userId]
	return ok
}

func newAllowedUserIds(userIds []string) map[string]struct{} {
	allowed := make(map[string]struct{})
	for _, userId := range userIds {
		userId = strings.TrimSpace(userId)
		if userId != "" {
			allowed[userId] = struct{}{}
		}
	}
	return allowed
}
