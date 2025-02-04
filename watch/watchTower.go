package watch

import (
	"time"
	"watch_bot/bots"
)

type Server struct {
	Name string
	URL  string
}

func Dog(server Server, messagesChannel chan bots.Message, chatId string, livenessChannel chan string, unhealthyThreshold int, checker URLChecker) {
	isAlive := true
	deadCounter := 0

	for message := range livenessChannel {
		if message != server.Name {
			continue
		}
		// if we think server is alive and it is really alive
		if isAlive && checker.IsUrlOk(server.URL) {
			// do nothing
			continue
		}
		// if we think server is alive, but it is not
		isAlive = isAlive || deadCounter == unhealthyThreshold
		deadCounter += 1
		if isAlive {
			// mark server as not alive
			isAlive = false
			// start goroutine to wait it to wake up
			go waitForWakeUp(server.URL, &isAlive, &deadCounter, messagesChannel, chatId, server.Name, checker)
			// notify about server is not OK
			messagesChannel <- bots.Message{ChatId: chatId, Text: server.Name + " is not OK"}
		}
	}
}

func waitForWakeUp(url string, isALive *bool, deadCounter *int, messagesChannel chan bots.Message, chatId string, name string, checker URLChecker) {
	for i := 0; i < 10; i++ {
		time.Sleep(5 * time.Second)
		if checker.IsUrlOk(url) {
			*isALive = true
			*deadCounter = 0
			messagesChannel <- bots.Message{ChatId: chatId, Text: name + " is OK now"}
			return
		}
	}
	*isALive = true
	*deadCounter = 0
	messagesChannel <- bots.Message{ChatId: chatId, Text: name + " is really not OK"}
}
