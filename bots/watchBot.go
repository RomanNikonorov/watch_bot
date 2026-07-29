package bots

import (
	"context"
	"time"
)

type WatchBot interface {
	CreateBot(context.Context, chan Command, string, chan Message, int, int, time.Duration) WatchBot
	ListenMessagesToSend(context.Context, chan Message, int, int)
	ListenIncomingMessages(context.Context, chan Command)
	Ready(context.Context) error
}
