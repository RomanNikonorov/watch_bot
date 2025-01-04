package bots

type WatchBot interface {
	CreateBot(string, chan Message) WatchBot
	ListenMessagesToSend(chan Message)
}
