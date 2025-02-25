package bots

type WatchBot interface {
	CreateBot(string, chan Message, int, int) WatchBot
	ListenMessagesToSend(chan Message, int, int)
}
