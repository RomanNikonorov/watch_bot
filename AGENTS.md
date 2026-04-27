# AGENTS.md

## Start Here

Before making changes, read this file first.

This repository is a Go service named `watch_bot`. It monitors server health, sends notifications through Telegram or VK Teams, exposes health/readiness/metrics endpoints, and supports the `\duty` bot command.

## First Checks

Use these as the default first steps:

- read `main.go`
- inspect `watch/`, `bots/`, `duty/`, `dao/`, `working_calendar/`
- run verification with `make test`

The project uses a local Go cache through `Makefile`. Prefer `make test`, `make build`, and `make run` over raw `go test` when sandboxed environments may block the system Go cache.

## Architecture

### Entry Point

Main flow is in `main.go`.

At startup the service:

1. reads configuration from environment variables;
2. initializes optional Graylog logging;
3. creates bot message and command channels;
4. loads monitored servers from PostgreSQL;
5. loads unusual working days from PostgreSQL;
6. initializes the selected bot (`telegram` or `vk`);
7. registers the `duty` command in the command router;
8. starts one watchdog goroutine per server;
9. starts an HTTP server on port `9000`;
10. schedules probes on a ticker during working time only;
11. handles graceful shutdown via `SIGINT` and `SIGTERM`.

### Bots

Implemented in `bots/`.

- `botFactory.go` selects bot implementation by `BOT_TYPE`.
- `telegramBot.go` handles:
  - incoming messages via Telegram updates;
  - command parsing;
  - outgoing messages with retry;
  - shutdown-aware loops using `context.Context`.
- `vkTeamsBot.go` provides the same for VK Teams.
- `retry.go` contains cancellation-aware retry waiting.

Incoming commands are accepted only from configured command chats. The `duty` command is restricted to `MAIN_CHAT_ID`; the `next` command is restricted to `SUPPORT_CHAT_ID` and sender user IDs listed in semicolon-separated `NEXT_ALLOWED_USER_IDS`.

### Command Routing

Implemented in `bots/commandRouter.go`.

- Commands must start with `\`.
- Parsed commands are normalized to lowercase.
- Unknown commands return a generated help message listing registered commands.
- The router currently registers two commands: `duty` and `next`.

### Duty Command

Implemented in:

- `bots/commands/duty.go`
- `duty/duty_service.go`
- DAO methods in `dao/dao.go`

Behavior of `\duty`:

- works only during configured working hours;
- loads duties from PostgreSQL;
- returns the current duty person for today if already assigned;
- otherwise selects the next person alphabetically after the last assigned one;
- if nobody has ever been assigned, selects the first alphabetically;
- updates `last_duty_date` when a new daily assignment is made;
- sends a direct message to the selected duty person;
- sends a support-chat notification on the first assignment of the day;
- returns a response to the caller chat: `The development team is rushing to help!`.

Support chat notification uses VK Teams mention format `@[userId]` and HTML parse mode.

### Next Command

Implemented in:

- `bots/commands/next.go`
- `duty/duty_service.go`
- DAO methods in `dao/dao.go`

Behavior of `\next`:

- works only during configured working hours;
- is accepted only from `SUPPORT_CHAT_ID`;
- is accepted only from users listed in semicolon-separated `NEXT_ALLOWED_USER_IDS`;
- returns a permission-denied response to other users;
- requires an existing duty assignment for today;
- selects the next person alphabetically after today's current duty person;
- clears today's `last_duty_date` from the current duty record and assigns today's date to the next duty record in a single transaction;
- sends a direct message to the newly selected duty person;
- sends a support-chat notification using VK Teams mention format `@[userId]` and HTML parse mode.

### Monitoring

Implemented in `watch/`.

- Servers are loaded from the `servers` table.
- Each server gets its own watchdog goroutine.
- Probe success requires HTTP status `200`.
- HTTPS certificate validation is disabled (`InsecureSkipVerify: true`).
- A server is considered failed only after the initial failed probe plus unhealthy retries.
- On first failure:
  - the bot sends a "not responding" notification;
  - the watchdog switches into recovery mode.
- While failed:
  - the watchdog probes using `DEAD_PROBE_DELAY`;
  - after `DEAD_PROBE_THRESHOLD` failures it sends one offline notification;
  - then sleeps for `DEAD_PROBE_PAUSE` minutes before retry cycles continue.
- On recovery:
  - the bot sends an "is responding" notification;
  - the watchdog returns to normal mode.

### Working Calendar

Implemented in `working_calendar/workingCalendar.go`.

Environment-driven schedule:

- `START_TIME`
- `END_TIME`
- `DAYS_OFF`

Additional exceptions are loaded from PostgreSQL table `unusual_days`.

Logic:

- if working-time config is invalid or absent, monitoring is effectively always enabled;
- configured days off are treated as non-working days;
- dates from `unusual_days` invert the normal behavior for that date.

### Database Access

Implemented in `dao/dao.go`.

Current tables used by the application:

- `servers`
- `unusual_days`
- `duties`

DAO responsibilities:

- load servers to monitor;
- load unusual days;
- load all duty records;
- update `last_duty_date` for duty rotation.

### HTTP Endpoints

Implemented in `main.go` and `lib/handlers.go`.

- `GET /health` -> `200 OK`
- `GET /ready` -> `200 OK` while ready, `503` during shutdown
- `GET /metrics` -> Prometheus metrics endpoint

`LoggerWithSkipPaths` skips access logging for `/health`, `/ready`, and `/metrics`.

### Shutdown Behavior

The service performs graceful shutdown:

- cancels the root context;
- stops probe scheduling;
- stops bot send/receive loops;
- stops command routing;
- marks readiness as failed;
- shuts down the HTTP server with a 10-second timeout.

## Configuration Summary

Important environment variables:

- `CONNECTION_STR`
- `BOT_TOKEN`
- `BOT_API_URL`
- `BOT_TYPE`
- `MAIN_CHAT_ID`
- `SUPPORT_CHAT_ID`
- `NEXT_ALLOWED_USER_IDS` (semicolon-separated)
- `PROBE_DELAY`
- `DEAD_PROBE_DELAY`
- `DEAD_PROBE_THRESHOLD`
- `DEAD_PROBE_PAUSE`
- `UNHEALTHY_THRESHOLD`
- `UNHEALTHY_DELAY`
- `PROBE_TIMEOUT`
- `RETRY_COUNT`
- `RETRY_PAUSE`
- `START_TIME`
- `END_TIME`
- `DAYS_OFF`
- `GRAYLOG_ADDR`

## Current Limitations

- `servers` and `unusual_days` are loaded only once at startup; DB changes require restart.
- `duties` are read dynamically on each `\duty` call.
- Monitoring accepts only HTTP `200` as healthy.
- TLS certificate verification is disabled.
- Commands are only accepted from configured command chats; each command has its own chat restriction.
- There is no admin UI or runtime reload mechanism.

## Verification

Default verification command:

- `make test`

This passed at the time this file was created.
