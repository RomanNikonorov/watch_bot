# AGENTS.md

## Start Here

Before making changes, read this file first.

This repository is a Go service named `watch_bot`. It sends duty notifications through Telegram or VK Teams, exposes health/readiness/metrics endpoints, and supports the `\duty` and `\next` bot commands.

## First Checks

Use these as the default first steps:

- check `git status --short` and preserve unrelated user changes
- read `main.go`
- inspect `bots/`, `duty/`, `dao/`, `working_calendar/`
- run verification with `make test`

The project uses a local Go cache through `Makefile`. Prefer `make test`, `make build`, and `make run` over raw `go test` when sandboxed environments may block the system Go cache.

The module currently declares Go `1.25.5` in `go.mod`. `make build` writes the binary to `.bin/watch_bot`; `.gocache/`, `.tmp/`, and `.bin/` are generated local artifacts and must not be committed.

## Change Guidelines

- Keep command parsing and transport-specific bot logic in `bots/`.
- Keep command orchestration and user-facing command responses in `bots/commands/`.
- Keep duty rotation rules in `duty/`; prefer pure functions such as `FindCurrentDuty` and `FindNextDuty` for logic that can be unit tested.
- Keep SQL and transaction boundaries in `dao/`.
- Keep schedule calculations in `working_calendar/`.
- Preserve cancellation-aware channel sends, receive loops, and retry waits when changing concurrent code.
- Add or update tests next to the changed package. Tests should use fixed timestamps where possible; avoid adding database or external bot dependencies to the default unit-test suite.
- Run `gofmt` on changed Go files, then `make test`. For runtime or packaging changes, also run `make build`.

## Architecture

### Entry Point

Main flow is in `main.go`.

At startup the service:

1. reads configuration from environment variables;
2. initializes optional Graylog logging;
3. validates typed configuration and the PostgreSQL connection, exiting on failure;
4. creates bot message and command channels;
5. loads unusual working days from PostgreSQL and periodically refreshes them;
6. initializes the selected bot (`telegram` or `vk`);
7. registers the `duty` and `next` commands in the command router;
8. starts an HTTP server on port `9000`;
9. handles graceful shutdown via `SIGINT` and `SIGTERM`.

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

`BOT_API_URL` is used by the VK Teams client. The Telegram client uses its library's configured API endpoint. Support-chat mentions are formatted per transport: `@[userId]` for VK Teams and an HTML `tg://user` link for Telegram.

### Command Routing

Implemented in `bots/commandRouter.go`.

- Commands must start with `\`.
- Parsed commands are normalized to lowercase.
- Remaining whitespace-separated tokens are stored as positional parameters under string keys (`"0"`, `"1"`, and so on).
- Unknown commands return a generated help message listing registered commands.
- The router registers `duty`; it registers `next` only when `SUPPORT_CHAT_ID` is non-empty.
- Chat restrictions are enforced both while receiving messages and by per-command handler wrappers.

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

Support chat notification uses the configured transport's mention format and HTML parse mode. Direct, support, and command-response messages wait for queue capacity or context cancellation; queue events and bot API attempts are exposed as Prometheus metrics.

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
- sends a support-chat notification using the configured transport's mention format and HTML parse mode.

If `NEXT_ALLOWED_USER_IDS` is empty, nobody is allowed to use `\next`. With fewer than two duty records, or without an assignment for the current day, no replacement is made.

### Working Calendar

Implemented in `working_calendar/workingCalendar.go`.

Environment-driven schedule:

- `START_TIME`
- `END_TIME`
- `DAYS_OFF`

Additional exceptions are loaded from PostgreSQL table `unusual_days` at startup and refreshed according to `UNUSUAL_DAYS_REFRESH_INTERVAL`.

Logic:

- if working-time config is invalid or absent, duty commands are effectively always enabled;
- configured days off are treated as non-working days;
- dates from `unusual_days` invert the normal behavior for that date.
- on a normal working day, both time bounds are inclusive;
- on an unusual configured day off, the whole day is treated as working without checking the time bounds;
- schedule calculations use the process-local timezone (`Local`/`TZ`).
- current duty dates and the unusual-day query use `lib.StartOfDay`, which preserves the calendar date and location instead of truncating by an absolute 24-hour duration.

### Database Access

Implemented in `dao/dao.go`.

Current tables used by the application:

- `unusual_days`
- `duties`

DAO responsibilities:

- load unusual days;
- load all duty records;
- update `last_duty_date` for duty rotation.

The application creates one shared `*sql.DB` pool. DAO operations accept `context.Context`. Duty selection and updates run inside one transaction that locks all duty rows with `SELECT ... FOR UPDATE`, serializing concurrent `\duty` and `\next` operations.

Versioned SQL migrations are in `migrations/` and are applied with `make migrate CONNECTION_STR=...`. They enforce unique `duty_id` values and at most one non-null `last_duty_date` per date. Migration `002_unique_duty_date.sql` intentionally fails if existing duplicate dates must be repaired first.

`duty_id` is both the alphabetical rotation key and the destination user/chat ID. The legacy `servers` table may still exist in deployments, but the application no longer reads it.

### HTTP Endpoints

Implemented in `main.go` and `lib/handlers.go`.

- `GET /health` -> `200 OK`
- `GET /ready` -> `200 OK` while ready and both PostgreSQL and the selected bot API respond, otherwise `503`
- `GET /metrics` -> Prometheus metrics endpoint

`LoggerWithSkipPaths` skips access logging for `/health`, `/ready`, and `/metrics`.

### Shutdown Behavior

The service performs graceful shutdown:

- cancels the root context;
- stops bot send/receive loops;
- stops command routing;
- marks readiness as failed;
- shuts down the HTTP server with a 10-second timeout.

## Configuration Summary

Important environment variables:

- `CONNECTION_STR` (required; startup validates it)
- `BOT_TOKEN` (required)
- `BOT_API_URL` (used by VK Teams)
- `BOT_TYPE` (`telegram` or `vk`; any other value terminates startup)
- `MAIN_CHAT_ID` (allowed chat for `\duty`)
- `SUPPORT_CHAT_ID` (notification chat and allowed chat for `\next`)
- `NEXT_ALLOWED_USER_IDS` (semicolon-separated)
- `RETRY_COUNT` (integer, default `3`)
- `RETRY_PAUSE` (integer seconds, default `5`)
- `DATABASE_TIMEOUT` (Go duration, default `5s`)
- `BOT_API_TIMEOUT` (Go duration, default `10s`)
- `UNUSUAL_DAYS_REFRESH_INTERVAL` (Go duration, default `5m`)
- `START_TIME` and `END_TIME` (`HH:MM`; both are needed to enable time checks)
- `DAYS_OFF` (comma-separated English weekday names, case-sensitive)
- `GRAYLOG_ADDR` (optional)
- `TZ` (sets the process timezone in the container)

## Current Limitations

- `duties` are read dynamically on every `\duty` and `\next` call.
- The outgoing queue is in-memory; messages can still be lost if the process exits or all bot API retries fail.
- Omitting both `START_TIME` and `END_TIME` intentionally disables working-hours checks; malformed or partial configuration fails startup validation.
- Commands are only accepted from configured command chats; each command has its own chat restriction.
- There is no admin UI or runtime reload mechanism.

## Verification

Default verification command:

```bash
make test
```

Useful additional checks:

```bash
make build
make vet
make race
make integration-test  # uses TEST_CONNECTION_STR when configured
```

The default test suite is package-level and does not require a live PostgreSQL instance or bot API. The build-tagged DAO integration test uses an isolated temporary PostgreSQL schema and is exercised by CI.
