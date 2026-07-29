package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ConnectionString           string
	BotToken                   string
	BotAPIURL                  string
	BotType                    string
	MainChatID                 string
	SupportChatID              string
	NextAllowedUserIDs         []string
	RetryCount                 int
	RetryPauseSeconds          int
	StartTime                  string
	EndTime                    string
	DaysOff                    string
	GraylogAddress             string
	DatabaseTimeout            time.Duration
	BotAPITimeout              time.Duration
	UnusualDaysRefreshInterval time.Duration
}

func Load() (Config, error) {
	return load(os.Getenv)
}

func load(getenv func(string) string) (Config, error) {
	retryCount, err := parseInt(getenv, "RETRY_COUNT", 3)
	if err != nil {
		return Config{}, err
	}
	if retryCount < 1 {
		return Config{}, fmt.Errorf("RETRY_COUNT must be at least 1")
	}

	retryPause, err := parseInt(getenv, "RETRY_PAUSE", 5)
	if err != nil {
		return Config{}, err
	}
	if retryPause < 0 {
		return Config{}, fmt.Errorf("RETRY_PAUSE must not be negative")
	}

	databaseTimeout, err := parseDuration(getenv, "DATABASE_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	botAPITimeout, err := parseDuration(getenv, "BOT_API_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	refreshInterval, err := parseDuration(getenv, "UNUSUAL_DAYS_REFRESH_INTERVAL", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		ConnectionString:           strings.TrimSpace(getenv("CONNECTION_STR")),
		BotToken:                   strings.TrimSpace(getenv("BOT_TOKEN")),
		BotAPIURL:                  strings.TrimSpace(getenv("BOT_API_URL")),
		BotType:                    strings.ToLower(strings.TrimSpace(getenv("BOT_TYPE"))),
		MainChatID:                 strings.TrimSpace(getenv("MAIN_CHAT_ID")),
		SupportChatID:              strings.TrimSpace(getenv("SUPPORT_CHAT_ID")),
		NextAllowedUserIDs:         splitSemicolonList(getenv("NEXT_ALLOWED_USER_IDS")),
		RetryCount:                 retryCount,
		RetryPauseSeconds:          retryPause,
		StartTime:                  strings.TrimSpace(getenv("START_TIME")),
		EndTime:                    strings.TrimSpace(getenv("END_TIME")),
		DaysOff:                    strings.TrimSpace(getenv("DAYS_OFF")),
		GraylogAddress:             strings.TrimSpace(getenv("GRAYLOG_ADDR")),
		DatabaseTimeout:            databaseTimeout,
		BotAPITimeout:              botAPITimeout,
		UnusualDaysRefreshInterval: refreshInterval,
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	switch {
	case c.ConnectionString == "":
		return fmt.Errorf("CONNECTION_STR is required")
	case c.BotToken == "":
		return fmt.Errorf("BOT_TOKEN is required")
	case c.BotType != "telegram" && c.BotType != "vk":
		return fmt.Errorf("BOT_TYPE must be either telegram or vk")
	case c.MainChatID == "":
		return fmt.Errorf("MAIN_CHAT_ID is required")
	case c.BotType == "vk" && c.BotAPIURL == "":
		return fmt.Errorf("BOT_API_URL is required for BOT_TYPE=vk")
	case (c.StartTime == "") != (c.EndTime == ""):
		return fmt.Errorf("START_TIME and END_TIME must be configured together")
	case c.DatabaseTimeout <= 0:
		return fmt.Errorf("DATABASE_TIMEOUT must be positive")
	case c.BotAPITimeout <= 0:
		return fmt.Errorf("BOT_API_TIMEOUT must be positive")
	case c.UnusualDaysRefreshInterval <= 0:
		return fmt.Errorf("UNUSUAL_DAYS_REFRESH_INTERVAL must be positive")
	}

	if c.StartTime != "" {
		if _, err := time.Parse("15:04", c.StartTime); err != nil {
			return fmt.Errorf("invalid START_TIME: %w", err)
		}
		if _, err := time.Parse("15:04", c.EndTime); err != nil {
			return fmt.Errorf("invalid END_TIME: %w", err)
		}
	}

	validDays := map[string]struct{}{
		"Sunday": {}, "Monday": {}, "Tuesday": {}, "Wednesday": {},
		"Thursday": {}, "Friday": {}, "Saturday": {},
	}
	for _, day := range strings.Split(c.DaysOff, ",") {
		day = strings.TrimSpace(day)
		if day == "" {
			continue
		}
		if _, ok := validDays[day]; !ok {
			return fmt.Errorf("invalid day in DAYS_OFF: %q", day)
		}
	}
	return nil
}

func parseInt(getenv func(string) string, name string, fallback int) (int, error) {
	value := strings.TrimSpace(getenv(name))
	if value == "" {
		return fallback, nil
	}
	result, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}
	return result, nil
}

func parseDuration(getenv func(string) string, name string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(getenv(name))
	if value == "" {
		return fallback, nil
	}
	result, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", name, err)
	}
	return result, nil
}

func splitSemicolonList(value string) []string {
	var result []string
	for _, item := range strings.Split(value, ";") {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}
