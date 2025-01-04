package main

import (
	"log"
	"os"
	"sync"
	"watch_bot/bots"
)

func main() {

	botToken := os.Getenv("BOT_TOKEN")
	botApiUrl := os.Getenv("BOT_API_URL")
	mainChatId := os.Getenv("MAIN_CHAT_ID")
	botType := os.Getenv("BOT_TYPE")

	messagesChannel := make(chan bots.Message)
	settings := bots.BotSettings{
		BotToken:        botToken,
		BotApiUrl:       botApiUrl,
		MainChatId:      mainChatId,
		BotType:         botType,
		MessagesChannel: messagesChannel,
	}

	connectionStr := os.Getenv("CONNECTION_STR")
	_, err := readFromDatabase(connectionStr)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		bots.ProduceBot(settings)
	}()

	messagesChannel <- bots.Message{ChatId: mainChatId, Text: "Hello, world!!!"}

	wg.Wait()

}
