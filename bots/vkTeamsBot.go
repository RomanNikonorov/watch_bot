package bots

import (
	"github.com/mail-ru-im/bot-golang"
	"log"
)

type VkTeamsBot struct {
	Bot       *botgolang.Bot
	BotApiUrl string
}

func (b VkTeamsBot) CreateBot(botToken string, messagesChannel chan Message) WatchBot {
	newBot, err := botgolang.NewBot(botToken, botgolang.BotApiURL(b.BotApiUrl))
	if err != nil {
		log.Fatal("wrong parameters for bot creation")
	}
	b.Bot = newBot
	go b.ListenMessagesToSend(messagesChannel)
	return b
}

func (b VkTeamsBot) ListenMessagesToSend(messagesChannel chan Message) {
	for message := range messagesChannel {
		botMessage := b.Bot.NewTextMessage(message.ChatId, message.Text)
		err := botMessage.Send()
		if err != nil {
			log.Printf("failed to send message: %v", err)
		}
	}
}
