package commands

import (
	"fmt"
	"log"
	"strings"
	"watch_bot/bots"
	"watch_bot/duty"

	botgolang "github.com/mail-ru-im/bot-golang"
)

type NextCommandConfig struct {
	ConnectionStr      string
	MessagesChan       chan bots.Message
	SupportChatId      string
	AllowedNextUserIds []string
	IsWorkingNow       func() bool
}

type nextDutyServicer interface {
	GetNextDuty() (*duty.DutyResult, error)
}

type NextCommand struct {
	dutyService        nextDutyServicer
	messagesChan       chan bots.Message
	supportChatId      string
	allowedNextUserIds map[string]struct{}
	isWorkingNow       func() bool
}

func NewNextCommand(config NextCommandConfig) *NextCommand {
	return &NextCommand{
		dutyService:        duty.NewService(config.ConnectionStr),
		messagesChan:       config.MessagesChan,
		supportChatId:      config.SupportChatId,
		allowedNextUserIds: newAllowedUserIds(config.AllowedNextUserIds),
		isWorkingNow:       config.IsWorkingNow,
	}
}

func (n *NextCommand) Execute(cmd bots.Command) (string, error) {
	if !n.isUserAllowed(cmd.UserId) {
		return "You are not allowed to execute this command", nil
	}

	if n.isWorkingNow != nil && !n.isWorkingNow() {
		return "Duty can only be changed during working hours", nil
	}

	result, err := n.dutyService.GetNextDuty()
	if err != nil {
		return "", fmt.Errorf("failed to get next duty: %w", err)
	}
	if result == nil {
		return "No next duty person available for today", nil
	}

	if n.messagesChan != nil {
		select {
		case n.messagesChan <- bots.Message{
			ChatId: result.DutyID,
			Text:   "You are on duty today!",
		}:
		default:
			log.Printf("Warning: failed to send notification to duty person %s: channel buffer full", result.DutyID)
		}

		if n.supportChatId != "" {
			notificationText := fmt.Sprintf("Duty person changed.\n\nOn duty now: @[%s]", result.DutyID)
			select {
			case n.messagesChan <- bots.Message{
				ChatId:    n.supportChatId,
				Text:      notificationText,
				ParseMode: string(botgolang.ParseModeHTML),
			}:
			default:
				log.Printf("Warning: failed to send notification to support chat %s: channel buffer full", n.supportChatId)
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
