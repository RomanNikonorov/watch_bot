package bots

import (
	"github.com/mail-ru-im/bot-golang"
	"log"
	"time"
)

type VkTeamsBot struct {
	Bot       *botgolang.Bot
	BotApiUrl string
}

func (b VkTeamsBot) CreateBot(botToken string, messagesChannel chan Message, retryCount int, retryPause int) WatchBot {
	newBot, err := botgolang.NewBot(botToken, botgolang.BotApiURL(b.BotApiUrl))
	if err != nil {
		log.Fatal("wrong parameters for bot creation")
	}
	b.Bot = newBot
	go b.ListenMessagesToSend(messagesChannel, retryCount, retryPause)
	return b
}

func (b VkTeamsBot) ListenMessagesToSend(messagesChannel chan Message, retryCount int, retryPause int) {
	for message := range messagesChannel {
		botMessage := b.Bot.NewTextMessage(message.ChatId, message.Text)
		vkSendWithRetry(botMessage.Send, retryCount, retryPause)
	}
}

func vkSendWithRetry(sendFunc func() error, retryCount int, retryPause int) {
	for i := 0; i < retryCount; i++ {
		err := sendFunc()
		if err != nil {
			log.Printf("failed to send message: %v in attempt %v", err, i)
			time.Sleep(time.Duration(retryPause) * time.Second)
		} else {
			break
		}
	}
}
