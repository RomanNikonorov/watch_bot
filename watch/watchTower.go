package watch

import (
	"log"
	"net/http"
	"time"
	"watch_bot/bots"
)

type Server struct {
	Name string
	URL  string
}

func Dog(server Server, messagesChannel chan bots.Message, chatId string, livenessChannel chan string, statusChannel chan LivenessStatus) {
	isAlive := true
	for message := range livenessChannel {
		if message != server.Name {
			continue
		}
		// if we think server is alive and it is really alive
		if isAlive && isUrlOk(server.URL) {
			// do nothing
			continue
		}
		// if we think server is alive, but it is not
		if isAlive {
			// mark server as not alive
			isAlive = false
			// start goroutine to wait it to wake up
			go waitForWakeUp(server.URL, &isAlive, messagesChannel, chatId, server.Name)
			// notify about server is not OK
			messagesChannel <- bots.Message{ChatId: chatId, Text: server.Name + " is not OK"}
		}
	}
}

func waitForWakeUp(url string, isALive *bool, messagesChannel chan bots.Message, chatId string, name string) {
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		if isUrlOk(url) {
			*isALive = true
			messagesChannel <- bots.Message{ChatId: chatId, Text: name + " is OK now"}
			return
		}
	}
	*isALive = true
	messagesChannel <- bots.Message{ChatId: chatId, Text: name + " is really not OK"}
}

func isUrlOk(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("failed to get URL: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

type LivenessStatus struct {
	ServerName string
	IsOk       bool
}
