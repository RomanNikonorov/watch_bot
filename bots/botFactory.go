package bots

import (
	"context"
	"log"
	"time"
)

func CreateBot(ctx context.Context, settings BotSettings) WatchBot {
	var bot WatchBot
	switch settings.BotType {
	case "vk":
		bot = &VkTeamsBot{}
		bot.(*VkTeamsBot).BotApiUrl = settings.BotApiUrl
		bot.(*VkTeamsBot).MainChatId = settings.MainChatId
		bot.(*VkTeamsBot).SupportChatId = settings.SupportChatId
	case "telegram":
		bot = &TelegramBot{}
		bot.(*TelegramBot).MainChatId = settings.MainChatId
		bot.(*TelegramBot).SupportChatId = settings.SupportChatId
	default:
		log.Fatal("unsupported bot type")
	}
	return bot.CreateBot(
		ctx,
		settings.CommandsChannel,
		settings.BotToken,
		settings.MessagesChannel,
		settings.RetryCount,
		settings.RetryPause,
		settings.APITimeout,
	)
}

type BotSettings struct {
	BotToken        string
	BotApiUrl       string
	MainChatId      string
	SupportChatId   string
	BotType         string
	MessagesChannel chan Message
	CommandsChannel chan Command
	RetryCount      int
	RetryPause      int
	APITimeout      time.Duration
}

type Message struct {
	ChatId    string
	Text      string
	ParseMode string // Transport-specific parse mode, such as HTML.
}

type Command struct {
	Name   string
	ChatId string
	UserId string
	Params map[string]string
}
