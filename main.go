package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"watch_bot/bots"
	"watch_bot/dao"
	"watch_bot/lib"
	"watch_bot/watch"
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
	botType := os.Getenv("BOT_TYPE")

	// delay between probes
	probeDelay := lib.GetEnvVariableValueWithDefault("PROBE_DELAY", "5")
	// delay between probes when server is dead
	deadProbeDelay := lib.GetEnvVariableValueWithDefault("DEAD_PROBE_DELAY", "60")
	// number of dead probes before sending a message
	deadThreshold := lib.GetEnvVariableValueWithDefault("DEAD_PROBE_THRESHOLD", "10")
	// pause in minutes before continuing to probe after server is dead
	deadPause := lib.GetEnvVariableValueWithDefault("DEAD_PROBE_PAUSE", "30")
	// number of unhealthy probes before sending a message
	unhealthyThreshold := lib.GetEnvVariableValueWithDefault("UNHEALTHY_THRESHOLD", "3")
	// delay between unhealthy probes
	unhealthyDelay := lib.GetEnvVariableValueWithDefault("UNHEALTHY_DELAY", "2")
	// retry count for bot
	retryCount := lib.GetEnvVariableValueWithDefault("RETRY_COUNT", "3")
	// retry pause for bot
	retryPause := lib.GetEnvVariableValueWithDefault("RETRY_PAUSE", "5")
	// probe timeout
	probeTimeout := lib.GetEnvVariableValueWithDefault("PROBE_TIMEOUT", "3")

	botMessagesChannel := make(chan bots.Message)
	botCommandsChannel := make(chan bots.Command)

	settings := bots.BotSettings{
		BotToken:        botToken,
		BotApiUrl:       botApiUrl,
		MainChatId:      mainChatId,
		BotType:         botType,
		MessagesChannel: botMessagesChannel,
		CommandsChannel: botCommandsChannel,
		RetryCount:      retryCount,
		RetryPause:      retryPause,
	}

	connectionStr := os.Getenv("CONNECTION_STR")
	servers, err := dao.GetServers(connectionStr)
	if err != nil {
		log.Fatal(err)
	}

	watchTowerLivenessChannelsMap := make(map[string]chan string)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bots.CreateBot(ctx, settings)
	for _, server := range servers {
		watchTowerLivenessChannelsMap[server.Name] = make(chan string)
		config := watch.DogConfig{
			Server:             server,
			LivenessChannel:    watchTowerLivenessChannelsMap[server.Name],
			MessagesChannel:    botMessagesChannel,
			UnhealthyThreshold: unhealthyThreshold,
			UnhealthyDelay:     unhealthyDelay,
			DeadProbeDelay:     deadProbeDelay,
			DeadThreshold:      deadThreshold,
			Checker:            watch.RealURLChecker{},
			ChatId:             mainChatId,
			DeadPause:          deadPause,
			ProbeTimeout:       probeTimeout,
		}
		go watch.Dog(config)
	}

	// graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		s := <-signalChan
		log.Printf("got signal %v, attempting graceful shutdown", s)
		cancel()
		wg.Done()
	}()

	// metrics & probes
	isReady := &atomic.Value{}
	isReady.Store(true)
	httpRouter := chi.NewRouter()
	httpRouter.Use(middleware.RequestID)
	httpRouter.Use(middleware.Logger)
	httpRouter.Use(middleware.Recoverer)

	httpRouter.HandleFunc("/health", lib.Healthz)
	httpRouter.HandleFunc("/ready", lib.Readyz(isReady))
	httpRouter.Handle("/metrics", promhttp.Handler())

	go func() {
		err := http.ListenAndServe(":9000", httpRouter)
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("Current time: %v", time.Now().Format("02.01.2006 MST"))
	workingCalendar := working_calendar.FillWorkingTime()
	unusualDays, err := dao.GetUnusualDays(connectionStr, time.Now())
	for _, day := range unusualDays {
		fmt.Printf("Unusual day: %s\n", day.Format("2006-01-02"))
	}
	if err != nil {
		log.Printf("Error getting unusual days: %v", err)
	}
	for {
		time.Sleep(time.Duration(probeDelay) * time.Second)
		for _, server := range servers {
			if working_calendar.IsWorkingTime(workingCalendar, time.Now(), unusualDays) {
				watchTowerLivenessChannelsMap[server.Name] <- server.Name
			}
		}
	}
}
