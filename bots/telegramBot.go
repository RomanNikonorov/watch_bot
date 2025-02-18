package bots

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
)

type TelegramBot struct {
	Bot *tgbotapi.BotAPI
}

func (b *TelegramBot) CreateBot(botToken string, messagesChannel chan Message) WatchBot {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal("wrong parameters for bot creation :", err)
	}
	b.Bot = bot
	go b.ListenMessagesToSend(messagesChannel)
	return b
}

func (b *TelegramBot) ListenMessagesToSend(messagesChannel chan Message) {
	for message := range messagesChannel {
		chatIdInt, err := strconv.ParseInt(message.ChatId, 10, 64)
		if err != nil {
			fmt.Println("Error:", err)
		}
		msg := tgbotapi.NewMessage(chatIdInt, message.Text)
		_, err = b.Bot.Send(msg)
		if err != nil {
			log.Printf("failed to send message: %v", err)
		}
	}
}
