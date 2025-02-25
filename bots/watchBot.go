package bots

type WatchBot interface {
	CreateBot(string, chan Message, int) WatchBot
	ListenMessagesToSend(chan Message, int)
}
