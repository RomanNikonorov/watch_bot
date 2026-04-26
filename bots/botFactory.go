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
		bot.(*VkTeamsBot).MainChatId = settings.MainChatId
		bot.(*VkTeamsBot).SupportChatId = settings.SupportChatId
	case "telegram":
		bot = &TelegramBot{}
		bot.(*TelegramBot).MainChatId = settings.MainChatId
		bot.(*TelegramBot).SupportChatId = settings.SupportChatId
	default:
		log.Fatal("unsupported bot type")
	}
	return bot.CreateBot(ctx, settings.CommandsChannel, settings.BotToken, settings.MessagesChannel, settings.RetryCount, settings.RetryPause)
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
}

type Message struct {
	ChatId    string
	Text      string
	ParseMode string // "HTML" or "MarkdownV2" for VK Teams
}

type Command struct {
	Name   string
	ChatId string
	UserId string
	Params map[string]string
}
