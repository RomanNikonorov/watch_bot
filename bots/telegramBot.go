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
	Bot        *tgbotapi.BotAPI
	MainChatId string
}

func (b *TelegramBot) ListenIncomingMessages(ctx context.Context, messages chan Command) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := b.Bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal("Failed to get updates channel:", err)
	}
	defer b.Bot.StopReceivingUpdates()

	for {
		select {
		case <-ctx.Done(): // Stop on context cancellation
			log.Println("Stopping ListenIncomingMessages:", ctx.Err())
			return
		case update, ok := <-updates: // Process incoming messages
			if !ok {
				return
			}
			if update.Message != nil {
				chatId := strconv.FormatInt(update.Message.Chat.ID, 10)
				// Only accept commands from main chat
				if b.MainChatId != "" && chatId != b.MainChatId {
					log.Printf("Ignoring message from chat %s (not main chat)", chatId)
					continue
				}
				log.Printf("Received message: %s from user id %d", update.Message.Text, update.Message.Chat.ID)
				cmd := ParseCommand(update.Message.Text, chatId)
				if cmd != nil {
					select {
					case <-ctx.Done():
						return
					case messages <- *cmd:
					}
				}
			}
		}
	}
}

func (b *TelegramBot) CreateBot(ctx context.Context, commandChannel chan Command, botToken string, messagesChannel chan Message, retryCount int, retryPause int) WatchBot {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal("wrong parameters for bot creation :", err)
	}
	b.Bot = bot
	go b.ListenMessagesToSend(ctx, messagesChannel, retryCount, retryPause)
	go b.ListenIncomingMessages(ctx, commandChannel)
	return b
}

func (b *TelegramBot) ListenMessagesToSend(ctx context.Context, messagesChannel chan Message, retryCount int, retryPause int) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping ListenMessagesToSend:", ctx.Err())
			return
		case message, ok := <-messagesChannel:
			if !ok {
				return
			}
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
			tgSendWithRetry(ctx, sendFunc, retryCount, retryPause)
		}
	}
}

func tgSendWithRetry(ctx context.Context, sendFunc func() error, retryCount, retryPause int) {
	for i := 0; i < retryCount; i++ {
		if ctx.Err() != nil {
			return
		}
		err := sendFunc()
		if err != nil {
			log.Printf("failed to send message: %v in attempt %v", err, i)
			if !waitForRetry(ctx, time.Duration(retryPause)*time.Second) {
				return
			}
		} else {
			break
		}
	}
}
