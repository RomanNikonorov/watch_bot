package bots

import (
	"context"
	"log"
	"time"

	"github.com/mail-ru-im/bot-golang"
)

type VkTeamsBot struct {
	Bot        *botgolang.Bot
	BotApiUrl  string
	MainChatId string
}

func (b VkTeamsBot) ListenIncomingMessages(ctx context.Context, messages chan Command) {
	updates := b.Bot.GetUpdatesChannel(ctx)
	for {
		select {
		case <-ctx.Done(): // Stop on context cancellation
			log.Println("Stopping ListenIncomingMessages:", ctx.Err())
			return
		case update, ok := <-updates: // Process incoming messages
			if !ok {
				return
			}
			chatId := update.Payload.Chat.ID
			// Only accept commands from main chat
			if b.MainChatId != "" && chatId != b.MainChatId {
				log.Printf("Ignoring message from chat %s (not main chat)", chatId)
				continue
			}
			log.Println("Received message:", update.Payload.Text)
			cmd := ParseCommand(update.Payload.Text, chatId)
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

func (b VkTeamsBot) CreateBot(ctx context.Context, commandChannel chan Command, botToken string, messagesChannel chan Message, retryCount int, retryPause int) WatchBot {
	newBot, err := botgolang.NewBot(botToken, botgolang.BotApiURL(b.BotApiUrl))
	if err != nil {
		log.Fatal("Bot is not created: ", err)
	}
	b.Bot = newBot
	go b.ListenMessagesToSend(ctx, messagesChannel, retryCount, retryPause)
	go b.ListenIncomingMessages(ctx, commandChannel)
	return b
}

func (b VkTeamsBot) ListenMessagesToSend(ctx context.Context, messagesChannel chan Message, retryCount int, retryPause int) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping ListenMessagesToSend:", ctx.Err())
			return
		case message, ok := <-messagesChannel:
			if !ok {
				return
			}
			botMessage := b.Bot.NewTextMessage(message.ChatId, message.Text)
			// Apply ParseMode if specified
			if message.ParseMode != "" {
				botMessage.AppendParseMode(botgolang.ParseMode(message.ParseMode))
			}
			vkSendWithRetry(ctx, botMessage.Send, retryCount, retryPause)
		}
	}
}

func vkSendWithRetry(ctx context.Context, sendFunc func() error, retryCount int, retryPause int) {
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
