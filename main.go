package main

import (
	"log"
	"os"
	"strconv"
	"time"
	"watch_bot/bots"
	"watch_bot/watch"
)

func main() {

	botToken := os.Getenv("BOT_TOKEN")
	botApiUrl := os.Getenv("BOT_API_URL")
	mainChatId := os.Getenv("MAIN_CHAT_ID")
	botType := os.Getenv("BOT_TYPE")

	probeDelayStr := os.Getenv("PROBE_DELAY")
	if probeDelayStr == "" {
		probeDelayStr = "5" // default value
	}
	probeDelay, err := strconv.Atoi(probeDelayStr)
	if err != nil {
		log.Fatal(err)
	}

	deadProbeDelayStr := os.Getenv("DEAD_PROBE_DELAY")
	if deadProbeDelayStr == "" {
		deadProbeDelayStr = "60" // default value
	}
	deadProbeDelay, err := strconv.Atoi(deadProbeDelayStr)
	if err != nil {
		log.Fatal(err)
	}

	deadThresholdStr := os.Getenv("DEAD_PROBE_DELAY")
	if deadThresholdStr == "" {
		deadThresholdStr = "10" // default value
	}
	deadThreshold, err := strconv.Atoi(deadThresholdStr)
	if err != nil {
		log.Fatal(err)
	}

	unhealthyThresholdStr := os.Getenv("UNHEALTHY_THRESHOLD")
	if unhealthyThresholdStr == "" {
		unhealthyThresholdStr = "3" // default value
	}
	unhealthyThreshold, err := strconv.Atoi(unhealthyThresholdStr)
	if err != nil {
		log.Fatal(err)
	}

	unhealthyDelayStr := os.Getenv("UNHEALTHY_DELAY")
	if unhealthyDelayStr == "" {
		unhealthyDelayStr = "2" // default value
	}
	unhealthyDelay, err := strconv.Atoi(unhealthyDelayStr)
	if err != nil {
		log.Fatal(err)
	}

	botMessagesChannel := make(chan bots.Message)
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
		config := watch.DogConfig{
			Server:             server,
			LivenessChannel:    watchTowerLivenessChannelsMap[server.Name],
			MessagesChannel:    botMessagesChannel,
			UnhealthyThreshold: unhealthyThreshold,
			UnhealthyDelay:     unhealthyDelay,
			DeadProbeDelay:     deadProbeDelay,
			DeadThreshold:      deadThreshold,
			Checker:            watch.RealURLChecker{},
			ChatId:             mainChatId,
		}
		go watch.Dog(config)
	}

	for {
		time.Sleep(time.Duration(probeDelay) * time.Second)
		for _, server := range servers {
			watchTowerLivenessChannelsMap[server.Name] <- server.Name
		}
	}
}
