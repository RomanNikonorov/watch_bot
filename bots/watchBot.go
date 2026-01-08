package bots

import "context"

type WatchBot interface {
	CreateBot(context.Context, chan Command, string, chan Message, int, int) WatchBot
	ListenMessagesToSend(chan Message, int, int)
	ListenIncomingMessages(context.Context, chan Command)
}
