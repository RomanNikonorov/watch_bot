package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
	"watch_bot/bots"
	"watch_bot/bots/commands"
	"watch_bot/config"
	"watch_bot/dao"
	"watch_bot/lib"
	"watch_bot/working_calendar"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/Graylog2/go-gelf.v1/gelf"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	// graylog
	if cfg.GraylogAddress != "" {
		gelfWriter, err := gelf.NewWriter(cfg.GraylogAddress)
		if err != nil {
			log.Fatalf("gelf.NewWriter: %s", err)
		}
		// log to both stderr and graylog
		log.SetOutput(io.MultiWriter(os.Stderr, gelfWriter))
		log.Printf("logging to stderr & graylog@'%s'", cfg.GraylogAddress)
	}

	botMessagesChannel := make(chan bots.Message, 100)
	botCommandsChannel := make(chan bots.Command)

	settings := bots.BotSettings{
		BotToken:        cfg.BotToken,
		BotApiUrl:       cfg.BotAPIURL,
		MainChatId:      cfg.MainChatID,
		SupportChatId:   cfg.SupportChatID,
		BotType:         cfg.BotType,
		MessagesChannel: botMessagesChannel,
		CommandsChannel: botCommandsChannel,
		RetryCount:      cfg.RetryCount,
		RetryPause:      cfg.RetryPauseSeconds,
		APITimeout:      cfg.BotAPITimeout,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbCtx, dbCancel := context.WithTimeout(ctx, cfg.DatabaseTimeout)
	store, err := dao.Open(dbCtx, cfg.ConnectionString)
	dbCancel()
	if err != nil {
		log.Fatalf("database validation failed: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("database close error: %v", err)
		}
	}()

	bot := bots.CreateBot(ctx, settings)

	log.Printf("Current time: %v", time.Now().Format("02.01.2006 MST"))
	workingCalendar := working_calendar.FillWorkingTime()
	calendarCtx, calendarCancel := context.WithTimeout(ctx, cfg.DatabaseTimeout)
	unusualDays, err := store.GetUnusualDays(calendarCtx, lib.StartOfDay(time.Now()))
	calendarCancel()
	if err != nil {
		log.Fatalf("failed to load unusual days: %v", err)
	}
	for _, day := range unusualDays {
		log.Printf("Unusual day: %s", day.Format("2006-01-02"))
	}
	unusualDaysValue := &atomic.Value{}
	unusualDaysValue.Store(unusualDays)
	go refreshUnusualDays(ctx, store, unusualDaysValue, cfg.UnusualDaysRefreshInterval, cfg.DatabaseTimeout)

	// Initialize command router
	commandRouter := bots.NewCommandRouter()
	isWorkingNow := func() bool {
		return working_calendar.IsWorkingTime(
			workingCalendar,
			time.Now(),
			unusualDaysValue.Load().([]time.Time),
		)
	}
	commandRouter.Register("duty", bots.NewChatRestrictedHandler(commands.NewDutyCommand(commands.DutyCommandConfig{
		Store:         store,
		MessagesChan:  botMessagesChannel,
		SupportChatId: settings.SupportChatId,
		BotType:       settings.BotType,
		IsWorkingNow:  isWorkingNow,
	}), settings.MainChatId))
	if settings.SupportChatId != "" {
		commandRouter.Register("next", bots.NewChatRestrictedHandler(commands.NewNextCommand(commands.NextCommandConfig{
			Store:              store,
			MessagesChan:       botMessagesChannel,
			SupportChatId:      settings.SupportChatId,
			BotType:            settings.BotType,
			AllowedNextUserIds: cfg.NextAllowedUserIDs,
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
	httpRouter.HandleFunc("/ready", lib.Readyz(isReady, func(ctx context.Context) error {
		pingCtx, pingCancel := context.WithTimeout(ctx, cfg.DatabaseTimeout)
		defer pingCancel()
		return store.Ping(pingCtx)
	}, func(ctx context.Context) error {
		botCtx, botCancel := context.WithTimeout(ctx, cfg.BotAPITimeout)
		defer botCancel()
		return bot.Ready(botCtx)
	}))
	httpRouter.Handle("/metrics", promhttp.Handler())

	httpServer := &http.Server{
		Addr:              ":9000",
		Handler:           httpRouter,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
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

func refreshUnusualDays(
	ctx context.Context,
	store *dao.Store,
	target *atomic.Value,
	interval time.Duration,
	timeout time.Duration,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			queryCtx, cancel := context.WithTimeout(ctx, timeout)
			days, err := store.GetUnusualDays(queryCtx, lib.StartOfDay(now))
			cancel()
			if err != nil {
				log.Printf("failed to refresh unusual days: %v", err)
				continue
			}
			target.Store(days)
			log.Printf("refreshed unusual days: %d entries", len(days))
		}
	}
}

func shutdownHTTPServer(httpServer *http.Server, isReady *atomic.Value) {
	isReady.Store(false)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
}
