package bots

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"time"
)

type TelegramBot struct {
	Bot *tgbotapi.BotAPI
}

func (b *TelegramBot) CreateBot(botToken string, messagesChannel chan Message, retryCount int, retryPause int) WatchBot {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal("wrong parameters for bot creation :", err)
	}
	b.Bot = bot
	go b.ListenMessagesToSend(messagesChannel, retryCount, retryPause)
	return b
}

func (b *TelegramBot) ListenMessagesToSend(messagesChannel chan Message, retryCount int, retryPause int) {
	for message := range messagesChannel {
		chatIdInt, err := strconv.ParseInt(message.ChatId, 10, 64)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		msg := tgbotapi.NewMessage(chatIdInt, message.Text)
		sendFunc := func() error {
			_, err := b.Bot.Send(msg)
			return err
		}
		tgSendWithRetry(sendFunc, retryCount, retryPause)
	}
}

func tgSendWithRetry(sendFunc func() error, retryCount, retryPause int) {
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
