package bots

import "log"

func CreateBot(settings BotSettings) {
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
	bot.CreateBot(settings.BotToken, settings.MessagesChannel)
}

type BotSettings struct {
	BotToken        string
	BotApiUrl       string
	MainChatId      string
	BotType         string
	MessagesChannel chan Message
}

type Message struct {
	ChatId string
	Text   string
}
