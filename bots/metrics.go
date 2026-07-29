package bots

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var outgoingQueueEvents = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "watch_bot",
		Name:      "outgoing_queue_events_total",
		Help:      "Outgoing message queue events by result.",
	},
	[]string{"result"},
)

var botSendAttempts = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "watch_bot",
		Name:      "send_attempts_total",
		Help:      "Bot API send attempts by transport and result.",
	},
	[]string{"transport", "result"},
)
