package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
	"watch_bot/bots"
	"watch_bot/bots/commands"
	"watch_bot/dao"
	"watch_bot/lib"
	"watch_bot/working_calendar"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/Graylog2/go-gelf.v1/gelf"
)

func main() {
	graylogAddr := os.Getenv("GRAYLOG_ADDR")
	// graylog
	if graylogAddr != "" {
		gelfWriter, err := gelf.NewWriter(graylogAddr)
		if err != nil {
			log.Fatalf("gelf.NewWriter: %s", err)
		}
		// log to both stderr and graylog
		log.SetOutput(io.MultiWriter(os.Stderr, gelfWriter))
		log.Printf("logging to stderr & graylog@'%s'", graylogAddr)
	}

	botToken := os.Getenv("BOT_TOKEN")
	botApiUrl := os.Getenv("BOT_API_URL")
	mainChatId := os.Getenv("MAIN_CHAT_ID")
	supportChatId := os.Getenv("SUPPORT_CHAT_ID")
	nextAllowedUserIds := parseSemicolonSeparatedList(os.Getenv("NEXT_ALLOWED_USER_IDS"))
	botType := os.Getenv("BOT_TYPE")

	// retry count for bot
	retryCount := lib.GetEnvVariableValueWithDefault("RETRY_COUNT", "3")
	// retry pause for bot
	retryPause := lib.GetEnvVariableValueWithDefault("RETRY_PAUSE", "5")

	botMessagesChannel := make(chan bots.Message, 100)
	botCommandsChannel := make(chan bots.Command)

	settings := bots.BotSettings{
		BotToken:        botToken,
		BotApiUrl:       botApiUrl,
		MainChatId:      mainChatId,
		SupportChatId:   supportChatId,
		BotType:         botType,
		MessagesChannel: botMessagesChannel,
		CommandsChannel: botCommandsChannel,
		RetryCount:      retryCount,
		RetryPause:      retryPause,
	}

	connectionStr := os.Getenv("CONNECTION_STR")
	if err := dao.ValidateConnection(connectionStr); err != nil {
		log.Fatalf("database validation failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bots.CreateBot(ctx, settings)

	log.Printf("Current time: %v", time.Now().Format("02.01.2006 MST"))
	workingCalendar := working_calendar.FillWorkingTime()
	unusualDays, err := dao.GetUnusualDays(connectionStr, time.Now())
	for _, day := range unusualDays {
		fmt.Printf("Unusual day: %s\n", day.Format("2006-01-02"))
	}
	if err != nil {
		log.Printf("Error getting unusual days: %v", err)
	}

	// Initialize command router
	commandRouter := bots.NewCommandRouter()
	isWorkingNow := func() bool {
		return working_calendar.IsWorkingTime(workingCalendar, time.Now(), unusualDays)
	}
	commandRouter.Register("duty", bots.NewChatRestrictedHandler(commands.NewDutyCommand(commands.DutyCommandConfig{
		ConnectionStr: connectionStr,
		MessagesChan:  botMessagesChannel,
		SupportChatId: settings.SupportChatId,
		IsWorkingNow:  isWorkingNow,
	}), settings.MainChatId))
	if settings.SupportChatId != "" {
		commandRouter.Register("next", bots.NewChatRestrictedHandler(commands.NewNextCommand(commands.NextCommandConfig{
			ConnectionStr:      connectionStr,
			MessagesChan:       botMessagesChannel,
			SupportChatId:      settings.SupportChatId,
			AllowedNextUserIds: nextAllowedUserIds,
			IsWorkingNow:       isWorkingNow,
		}), settings.SupportChatId))
	}
	go commandRouter.Listen(ctx, botCommandsChannel, botMessagesChannel)

	isReady := &atomic.Value{}
	isReady.Store(true)
	httpRouter := chi.NewRouter()
	httpRouter.Use(middleware.RequestID)
	httpRouter.Use(lib.LoggerWithSkipPaths("/health", "/ready", "/metrics"))
	httpRouter.Use(middleware.Recoverer)

	httpRouter.HandleFunc("/health", lib.Healthz)
	httpRouter.HandleFunc("/ready", lib.Readyz(isReady))
	httpRouter.Handle("/metrics", promhttp.Handler())

	httpServer := &http.Server{
		Addr:    ":9000",
		Handler: httpRouter,
	}

	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signalChan)

	go func() {
		select {
		case <-ctx.Done():
		case s := <-signalChan:
			log.Printf("got signal %v, attempting graceful shutdown", s)
			cancel()
		}
	}()

	<-ctx.Done()
	shutdownHTTPServer(httpServer, isReady)
}

func shutdownHTTPServer(httpServer *http.Server, isReady *atomic.Value) {
	isReady.Store(false)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
}

func parseSemicolonSeparatedList(value string) []string {
	var result []string
	for _, item := range strings.Split(value, ";") {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}
