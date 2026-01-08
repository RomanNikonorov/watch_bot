package commands

import (
	"fmt"
	"watch_bot/bots"
	"watch_bot/duty"
)

// DutyCommand handles the \duty command
type DutyCommand struct {
	dutyService *duty.Service
}

// NewDutyCommand creates a new DutyCommand
func NewDutyCommand(connectionStr string) *DutyCommand {
	return &DutyCommand{
		dutyService: duty.NewService(connectionStr),
	}
}

// Execute handles the duty command
func (d *DutyCommand) Execute(cmd bots.Command) (string, error) {
	currentDuty, err := d.dutyService.GetCurrentDuty()
	if err != nil {
		return "", fmt.Errorf("failed to get current duty: %w", err)
	}
	if currentDuty == nil {
		return "Дежурный на сегодня не назначен", nil
	}
	return fmt.Sprintf("Сегодня дежурит: %s", currentDuty.DutyID), nil
}

// Description returns command description
func (d *DutyCommand) Description() string {
	return "показать текущего дежурного"
}
