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

	// delay between probes
	probeDelay := GetEnvVariableValueWithDefault("PROBE_DELAY", "5")
	// delay between probes when server is dead
	deadProbeDelay := GetEnvVariableValueWithDefault("DEAD_PROBE_DELAY", "60")
	// number of dead probes before sending a message
	deadThreshold := GetEnvVariableValueWithDefault("DEAD_PROBE_THRESHOLD", "10")
	// number of unhealthy probes before sending a message
	unhealthyThreshold := GetEnvVariableValueWithDefault("UNHEALTHY_THRESHOLD", "3")
	// delay between unhealthy probes
	unhealthyDelay := GetEnvVariableValueWithDefault("UNHEALTHY_DELAY", "2")

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
