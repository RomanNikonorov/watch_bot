package watch

import (
	"strconv"
	"time"
	"watch_bot/bots"
)

type Server struct {
	Name string
	URL  string
}

type DogConfig struct {
	Server             Server
	LivenessChannel    chan string
	MessagesChannel    chan bots.Message
	UnhealthyThreshold int
	UnhealthyDelay     int
	DeadProbeDelay     int
	Checker            URLChecker
	ChatId             string
}

func Dog(config DogConfig) {
	isAlive := true
	deadCounter := 0

	for message := range config.LivenessChannel {
		if message != config.Server.Name {
			continue
		}
		// if we think server is alive and it is really alive
		if isAlive && config.Checker.IsUrlOk(config.Server.URL, config.UnhealthyThreshold, config.UnhealthyDelay) {
			// do nothing
			continue
		}
		// if we think server is alive, but it is not
		if isAlive {
			if deadCounter < config.UnhealthyThreshold {
				// increase dead counter
				deadCounter++
				continue
			}
			// mark server as not alive
			isAlive = false
			// start goroutine to wait it to wake up
			go waitForWakeUp(config, &isAlive, &deadCounter)
			// notify about server is not OK
			config.MessagesChannel <- bots.Message{ChatId: config.ChatId, Text: "❌ " + config.Server.Name + " is not responding ❌"}
		}
	}
}

func waitForWakeUp(config DogConfig, isALive *bool, deadCounter *int) {
	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(config.DeadProbeDelay) * time.Second)
		if config.Checker.IsUrlOk(config.Server.URL, config.UnhealthyThreshold, config.UnhealthyDelay) {
			*isALive = true
			*deadCounter = 0
			config.MessagesChannel <- bots.Message{ChatId: config.ChatId, Text: "✅ " + config.Server.Name + " is back online ✅"}
			return
		}
	}
	config.MessagesChannel <- bots.Message{ChatId: config.ChatId, Text: "❌❌❌ " + config.Server.Name + " is really not OK, pause for " + strconv.Itoa(config.DeadProbeDelay) + " minutes ❌❌❌"}
	time.Sleep(time.Duration(config.DeadProbeDelay) * time.Minute)
	*isALive = true
	*deadCounter = 0
}
