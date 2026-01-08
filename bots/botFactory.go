package bots

import (
	"context"
	"log"
)

func CreateBot(ctx context.Context, settings BotSettings) WatchBot {
	var bot WatchBot
	switch settings.BotType {
	case "vk":
		bot = &VkTeamsBot{}
		bot.(*VkTeamsBot).BotApiUrl = settings.BotApiUrl
	case "telegram":
		bot = &TelegramBot{}
	default:
		log.Fatal("unsupported bot type")
	}
	return bot.CreateBot(ctx, settings.CommandsChannel, settings.BotToken, settings.MessagesChannel, settings.RetryCount, settings.RetryPause)
}

type BotSettings struct {
	BotToken        string
	BotApiUrl       string
	MainChatId      string
	BotType         string
	MessagesChannel chan Message
	CommandsChannel chan Command
	RetryCount      int
	RetryPause      int
}

type Message struct {
	ChatId string
	Text   string
}

type Command struct {
	Name   string
	ChatId string
	Params map[string]string
}
