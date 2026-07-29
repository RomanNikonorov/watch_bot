package bots

import (
	"context"
	"fmt"
	"html"
)

func runReadinessCheck(ctx context.Context, check func() error) error {
	result := make(chan error, 1)
	go func() {
		result <- check()
	}()

	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// EnqueueMessage waits until the outgoing queue accepts a message or ctx is canceled.
func EnqueueMessage(ctx context.Context, messages chan Message, message Message) error {
	if messages == nil {
		return nil
	}
	select {
	case messages <- message:
		outgoingQueueEvents.WithLabelValues("enqueued").Inc()
		return nil
	case <-ctx.Done():
		outgoingQueueEvents.WithLabelValues("canceled").Inc()
		return ctx.Err()
	}
}

// FormatMention returns transport-specific mention text and parse mode.
func FormatMention(botType, userID string) (string, string) {
	switch botType {
	case "telegram":
		return fmt.Sprintf(
			`<a href="tg://user?id=%s">duty person</a>`,
			html.EscapeString(userID),
		), "HTML"
	default:
		return fmt.Sprintf("@[%s]", userID), "HTML"
	}
}
