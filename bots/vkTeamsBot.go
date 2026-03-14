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
		case update := <-updates: // Process incoming messages
			chatId := update.Payload.Chat.ID
			// Only accept commands from main chat
			if b.MainChatId != "" && chatId != b.MainChatId {
				log.Printf("Ignoring message from chat %s (not main chat)", chatId)
				continue
			}
			log.Println("Received message:", update.Payload.Text)
			cmd := ParseCommand(update.Payload.Text, chatId)
			if cmd != nil {
				messages <- *cmd
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
	go b.ListenMessagesToSend(messagesChannel, retryCount, retryPause)
	go b.ListenIncomingMessages(ctx, commandChannel)
	return b
}

func (b VkTeamsBot) ListenMessagesToSend(messagesChannel chan Message, retryCount int, retryPause int) {
	for message := range messagesChannel {
		botMessage := b.Bot.NewTextMessage(message.ChatId, message.Text)
		// Apply ParseMode if specified
		if message.ParseMode != "" {
			botMessage.AppendParseMode(botgolang.ParseMode(message.ParseMode))
		}
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
