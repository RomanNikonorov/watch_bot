package main

import (
	"log"
	"os"
	"time"
	"watch_bot/bots"
	"watch_bot/watch"
)

func main() {

	botToken := os.Getenv("BOT_TOKEN")
	botApiUrl := os.Getenv("BOT_API_URL")
	mainChatId := os.Getenv("MAIN_CHAT_ID")
	botType := os.Getenv("BOT_TYPE")

	botMessagesChannel := make(chan bots.Message)
	watchDogStatusChannel := make(chan watch.LivenessStatus)
	settings := bots.BotSettings{
		BotToken:        botToken,
		BotApiUrl:       botApiUrl,
		MainChatId:      mainChatId,
		BotType:         botType,
		MessagesChannel: botMessagesChannel,
	}

	connectionStr := os.Getenv("CONNECTION_STR")
	servers, err := getServers(connectionStr)
	if err != nil {
		log.Fatal(err)
	}

	watchTowerLivenessChannelsMap := make(map[string]chan string)
	bots.CreateBot(settings)
	botMessagesChannel <- bots.Message{ChatId: mainChatId, Text: "WatchBot is on duty"}
	for _, server := range servers {
		watchTowerLivenessChannelsMap[server.Name] = make(chan string)
		go watch.Dog(server, botMessagesChannel, mainChatId, watchTowerLivenessChannelsMap[server.Name], watchDogStatusChannel)
	}

	for {
		time.Sleep(1 * time.Second)
		for _, server := range servers {
			watchTowerLivenessChannelsMap[server.Name] <- server.Name
		}
	}
}
