package bots

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRunReadinessCheckHonorsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	release := make(chan struct{})
	defer close(release)

	err := runReadinessCheck(ctx, func() error {
		<-release
		return errors.New("late result")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestEnqueueMessageStopsWhenContextIsCanceled(t *testing.T) {
	messages := make(chan Message)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := EnqueueMessage(ctx, messages, Message{ChatId: "chat", Text: "text"})
	if err == nil {
		t.Fatal("expected canceled enqueue to return an error")
	}
}

func TestFormatMentionByTransport(t *testing.T) {
	vkMention, vkMode := FormatMention("vk", "user-1")
	if vkMention != "@[user-1]" || vkMode != "HTML" {
		t.Fatalf("unexpected VK mention: %q, %q", vkMention, vkMode)
	}

	telegramMention, telegramMode := FormatMention("telegram", `1"><script>`)
	if telegramMode != "HTML" {
		t.Fatalf("expected Telegram HTML parse mode, got %q", telegramMode)
	}
	if !strings.Contains(telegramMention, "tg://user?id=") {
		t.Fatalf("unexpected Telegram mention: %q", telegramMention)
	}
	if strings.Contains(telegramMention, "<script>") {
		t.Fatalf("expected escaped Telegram user ID, got %q", telegramMention)
	}
}
