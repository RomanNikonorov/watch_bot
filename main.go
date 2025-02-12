package main

import (
	"context"
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
	"watch_bot/watch"

	"watch_bot/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/Graylog2/go-gelf.v1/gelf"
)

func main() {
	graylogAddr := os.Getenv("GRAYLOG_ADDR")
	// greylog
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
	probeDelay := GetEnvVariableValueWithDefault("PROBE_DELAY", "5")
	// delay between probes when server is dead
	deadProbeDelay := GetEnvVariableValueWithDefault("DEAD_PROBE_DELAY", "60")
	// number of dead probes before sending a message
	deadThreshold := GetEnvVariableValueWithDefault("DEAD_PROBE_THRESHOLD", "10")
	// pause in minutes before continuing to probe after server is dead
	deadPause := GetEnvVariableValueWithDefault("DEAD_PROBE_PAUSE", "30")
	// number of unhealthy probes before sending a message
	unhealthyThreshold := GetEnvVariableValueWithDefault("UNHEALTHY_THRESHOLD", "3")
	// delay between unhealthy probes
	unhealthyDelay := GetEnvVariableValueWithDefault("UNHEALTHY_DELAY", "2")

	botMessagesChannel := make(chan bots.Message)
	settings := bots.BotSettings{
		BotToken:        botToken,
		BotApiUrl:       botApiUrl,
		MainChatId:      mainChatId,
		BotType:         botType,
		MessagesChannel: botMessagesChannel,
	}

	connectionStr := os.Getenv("CONNECTION_STR")
	servers, err := getServers(connectionStr)
	if err != nil {
		log.Fatal(err)
	}

	watchTowerLivenessChannelsMap := make(map[string]chan string)
	bots.CreateBot(settings)
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
		}
		go watch.Dog(config)
	}

	// graceful shutdown
	_, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		s := <-sigCh
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

	httpRouter.HandleFunc("/health", handlers.Healthz)
	httpRouter.HandleFunc("/ready", handlers.Readyz(isReady))
	httpRouter.Handle("/metrics", promhttp.Handler())

	go func() {
		err := http.ListenAndServe(":9000", httpRouter)
		if err != nil {
			log.Fatal(err)
		}
	}()

	for {
		time.Sleep(time.Duration(probeDelay) * time.Second)
		for _, server := range servers {
			watchTowerLivenessChannelsMap[server.Name] <- server.Name
		}
	}
}
