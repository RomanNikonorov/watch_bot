package bots

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type TelegramBot struct {
	Bot *tgbotapi.BotAPI
}

func (b *TelegramBot) ListenIncomingMessages(ctx context.Context, messages chan Command) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := b.Bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal("Failed to get updates channel:", err)
	}
	for {
		select {
		case <-ctx.Done(): // Завершаем работу при отмене контекста
			log.Println("Stopping ListenIncomingMessages:", ctx.Err())
			return
		case update := <-updates: // Обрабатываем входящие сообщения
			if update.Message != nil {
				log.Printf("Received message: %s from user id %d", update.Message.Text, update.Message.Chat.ID)
			}
		}
	}
}

func (b *TelegramBot) CreateBot(tx context.Context, commandChannel chan Command, botToken string, messagesChannel chan Message, retryCount int, retryPause int) WatchBot {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal("wrong parameters for bot creation :", err)
	}
	b.Bot = bot
	go b.ListenMessagesToSend(messagesChannel, retryCount, retryPause)
	go b.ListenIncomingMessages(context.Background(), commandChannel)
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
