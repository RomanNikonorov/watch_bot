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
		return "Дежурный на сегодня не назначен", nil
	}

	// Отправляем уведомление дежурному через канал
	if d.messagesChan != nil {
		d.messagesChan <- bots.Message{
			ChatId: result.DutyID,
			Text:   "Сегодня ты дежурный!",
		}
	}

	return "Команда разработки спешит на помощь!", nil
}

// Description returns command description
func (d *DutyCommand) Description() string {
	return "показать текущего дежурного"
}
