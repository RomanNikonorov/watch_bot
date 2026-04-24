package watch

import (
	"context"
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
	ProbeTimeout       int
	Checker            URLChecker
	ChatId             string
}

func Dog(ctx context.Context, config DogConfig) {
	isAlive := true

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   time.Duration(config.ProbeTimeout) * time.Second,
		Transport: tr,
	}

	waitChannel := make(chan bool)

	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-config.LivenessChannel:
			if !ok {
				return
			}
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
				go waitForWakeUp(ctx, config, waitChannel, client)
				// notify about server is not OK
				sendMessage(ctx, config.MessagesChannel, bots.Message{ChatId: config.ChatId, Text: "❌ " + config.Server.Name + " is not responding ❌"})
			}
		case waitMessage := <-waitChannel:
			if waitMessage {
				isAlive = true
			}
		}
	}
}

func waitForWakeUp(ctx context.Context, config DogConfig, waitChan chan bool, client HTTPClient) {

	log.Printf("Start waiting for server %s to wake up with %d probes %d seconds each", config.Server.Name, config.DeadThreshold, config.DeadProbeDelay)
	for i := 0; i < config.DeadThreshold; i++ {
		if !waitForDuration(ctx, time.Duration(config.DeadProbeDelay)*time.Second) {
			return
		}
		if checkAndReport(ctx, config, client, waitChan) {
			return
		}
		log.Printf("Server %s is still dead after %d probes", config.Server.Name, i+1)
	}
	pauseMinutes := config.DeadPause
	sendMessage(ctx, config.MessagesChannel, bots.Message{ChatId: config.ChatId, Text: "🆘☠️ " + config.Server.Name + " is offline, pause watching it for " + strconv.Itoa(pauseMinutes) + " minutes ☠️🆘"})
	if !waitForDuration(ctx, time.Duration(pauseMinutes)*time.Minute) {
		return
	}
	checkAndReport(ctx, config, client, waitChan)
}

func checkAndReport(ctx context.Context, config DogConfig, client HTTPClient, waitChan chan bool) bool {
	isOk := config.Checker.IsUrlOk(config.Server.URL, config.UnhealthyThreshold, config.UnhealthyDelay, client)
	if isOk {
		sendMessage(ctx, config.MessagesChannel, bots.Message{ChatId: config.ChatId, Text: "✅ " + config.Server.Name + " is responding ✅"})
		select {
		case <-ctx.Done():
			return false
		case waitChan <- true:
		}
	}
	return isOk
}

func waitForDuration(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func sendMessage(ctx context.Context, messages chan bots.Message, message bots.Message) {
	select {
	case <-ctx.Done():
	case messages <- message:
	}
}
