package bots

import (
	"github.com/mail-ru-im/bot-golang"
	"log"
)

type VkTeamsBot struct {
	Bot       *botgolang.Bot
	BotApiUrl string
}

func (b VkTeamsBot) CreateBot(botToken string, messagesChannel chan Message, retryCount int) WatchBot {
	newBot, err := botgolang.NewBot(botToken, botgolang.BotApiURL(b.BotApiUrl))
	if err != nil {
		log.Fatal("wrong parameters for bot creation")
	}
	b.Bot = newBot
	go b.ListenMessagesToSend(messagesChannel, retryCount)
	return b
}

func (b VkTeamsBot) ListenMessagesToSend(messagesChannel chan Message, retryCount int) {
	for message := range messagesChannel {
		botMessage := b.Bot.NewTextMessage(message.ChatId, message.Text)
		err := botMessage.Send()
		for i := 0; i < retryCount; i++ {
			if err != nil {
				log.Printf("failed to send message: %v in attempt %v", err, i)
				err = botMessage.Send()
			} else {
				break
			}
		}
	}
}
