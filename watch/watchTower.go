package watch

import (
	"crypto/tls"
	"log"
	"net/http"
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
	DeadThreshold      int
	DeadPause          int
	Checker            URLChecker
	ChatId             string
}

func Dog(config DogConfig) {
	isAlive := true

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   time.Duration(1) * time.Second,
		Transport: tr,
	}

	waitChannel := make(chan bool)

	for {
		select {
		case message := <-config.LivenessChannel:
			if message != config.Server.Name {
				continue
			}
			// if we think server is alive and it is really alive
			isNowOk := config.Checker.IsUrlOk(config.Server.URL, config.UnhealthyThreshold, config.UnhealthyDelay, client)
			if isAlive && isNowOk {
				// do nothing
				continue
			}
			// if we think server is alive, but it is not
			if isAlive {
				// mark server as not alive
				isAlive = false
				// start goroutine to wait it to wake up
				go waitForWakeUp(config, waitChannel, client)
				// notify about server is not OK
				config.MessagesChannel <- bots.Message{ChatId: config.ChatId, Text: "❌ " + config.Server.Name + " is not responding ❌"}
			}
		case waitMessage := <-waitChannel:
			if waitMessage {
				isAlive = true
			}
		}
	}
}

func waitForWakeUp(config DogConfig, waitChan chan bool, client HTTPClient) {

	log.Printf("Start waiting for server %s to wake up with %d probes %d seconds each", config.Server.Name, config.DeadThreshold, config.DeadProbeDelay)
	for i := 0; i < config.DeadThreshold; i++ {
		time.Sleep(time.Duration(config.DeadProbeDelay) * time.Second)
		if checkAndReport(config, client, waitChan) {
			return
		}
		log.Printf("Server %s is still dead after %d probes", config.Server.Name, i+1)
	}
	pauseMinutes := config.DeadPause
	config.MessagesChannel <- bots.Message{ChatId: config.ChatId, Text: "☠️ " + config.Server.Name + " is offline, pause watching it for " + strconv.Itoa(pauseMinutes) + " minutes ☠️"}
	time.Sleep(time.Duration(pauseMinutes) * time.Minute)
	checkAndReport(config, client, waitChan)
}

func checkAndReport(config DogConfig, client HTTPClient, waitChan chan bool) bool {
	isOk := config.Checker.IsUrlOk(config.Server.URL, config.UnhealthyThreshold, config.UnhealthyDelay, client)
	if isOk {
		config.MessagesChannel <- bots.Message{ChatId: config.ChatId, Text: "✅ " + config.Server.Name + " is back online ✅"}
		waitChan <- true
	}
	return isOk
}
