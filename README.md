# WatchBot

WatchBot is a duty bot service for Telegram or VK Teams. It exposes health/readiness/metrics endpoints and supports the `\\duty` and `\\next` commands for daily duty rotation.

## Local Development

The repository includes a `Makefile` that pins Go build artifacts to local directories inside the project. This avoids failures in restricted environments where the default Go cache is not writable.

```bash
make test
make vet
make race
make build
make run
```

`make build` writes the binary to `.bin/watch_bot`. Temporary Go build artifacts are stored in `.gocache` and `.tmp`, so local runs do not depend on a writable system Go cache.

## Runtime Endpoints

The service starts an HTTP server on port `9000` with the following endpoints:

- `GET /health` - liveness probe
- `GET /ready` - readiness probe, including PostgreSQL and bot API connectivity
- `GET /metrics` - Prometheus metrics

## Graceful Shutdown

The process listens for `SIGINT` and `SIGTERM`.

On shutdown it:

- cancels the main application context
- stops bot send/receive loops
- interrupts bot retry pauses
- marks readiness as failed
- gracefully stops the HTTP server with a 10-second timeout

## Environment Variables

### Database Configuration
- `CONNECTION_STR`: PostgreSQL connection string (format: `postgres://username:password@host:port/dbname?sslmode=disable`)

### Bot Configuration
- `BOT_TOKEN`: Bot token
- `BOT_API_URL`: Bot API URL
- `BOT_API_TIMEOUT`: Bot API and readiness timeout as a Go duration (default: `10s`)
- `MAIN_CHAT_ID`: Main chat ID for notifications
- `SUPPORT_CHAT_ID`: Support chat ID for duty notifications and the `\\next` command (required for duty replacement)
- `NEXT_ALLOWED_USER_IDS`: Semicolon-separated list of user IDs allowed to execute `\\next`
- `BOT_TYPE`: Type of bot to use (can be `telegram` or `vk`)
- `RETRY_COUNT`: Number of attempts to send a message (default: 3)
- `RETRY_PAUSE`: Pause between retry attempts in seconds (default: 5)
- `DATABASE_TIMEOUT`: PostgreSQL operation timeout as a Go duration (default: `5s`)

### Working Calendar Configuration
- `START_TIME`: Start of working hours (format: "HH:MM", e.g., "09:00")
- `END_TIME`: End of working hours (format: "HH:MM", e.g., "18:00")
- `DAYS_OFF`: Comma-separated list of days off (e.g., "Saturday,Sunday")
- `UNUSUAL_DAYS_REFRESH_INTERVAL`: Refresh interval for database calendar exceptions (default: `5m`)

### Logging
- `GRAYLOG_ADDR`: Graylog server address (optional)

## How To Run

1. Apply the versioned database migrations:

```bash
make migrate CONNECTION_STR='postgres://username:password@localhost:5432/watch_bot?sslmode=disable'
```

Migration `002_unique_duty_date.sql` fails safely when existing data contains duplicate non-null `last_duty_date` values. Repair those rows before retrying the migration.

The complete migration SQL, in application order, is included below so it can also be executed by a Kubernetes Job, a deployment pipeline, or a database administrator.

### Migration 001: Initial Schema

```sql
CREATE TABLE IF NOT EXISTS unusual_days (
    id bigserial PRIMARY KEY,
    unusual_date date NOT NULL
);

CREATE TABLE IF NOT EXISTS duties (
    id bigserial PRIMARY KEY,
    duty_id text NOT NULL,
    last_duty_date date
);

CREATE UNIQUE INDEX IF NOT EXISTS duties_duty_id_uidx
    ON duties (duty_id);
```

### Migration 002: One Duty Person Per Date

```sql
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM duties
        WHERE last_duty_date IS NOT NULL
        GROUP BY last_duty_date
        HAVING count(*) > 1
    ) THEN
        RAISE EXCEPTION
            'cannot enforce one duty per date: duplicate last_duty_date values exist';
    END IF;
END
$$;

CREATE UNIQUE INDEX IF NOT EXISTS duties_last_duty_date_uidx
    ON duties (last_duty_date)
    WHERE last_duty_date IS NOT NULL;
```

2. Set the required environment variables. Configuration is validated before external services start. Minimal example for Telegram:

```bash
export CONNECTION_STR='postgres://username:password@localhost:5432/watch_bot?sslmode=disable'
export BOT_TYPE='telegram'
export BOT_TOKEN='your-bot-token'
export MAIN_CHAT_ID='your-main-chat-id'
export SUPPORT_CHAT_ID='your-support-chat-id'
export NEXT_ALLOWED_USER_IDS='user-id-1;user-id-2'

export RETRY_COUNT='3'
export RETRY_PAUSE='5'
```

3. Start the service:

```bash
make run
```

4. Verify that the process is up:

```bash
curl http://localhost:9000/health
curl http://localhost:9000/ready
```

## Bot Commands

`\\duty` shows the current duty person. It is accepted from `MAIN_CHAT_ID`. When called, the bot returns a message indicating help is on the way, notifies the person on duty, and on the first assignment of the day also sends a notification to the support chat. For VK Teams, that support notification is sent with HTML parse mode.

`\\next` replaces today's duty person with the next person in alphabetical rotation. It is accepted from `SUPPORT_CHAT_ID` only when the sender user ID is listed in `NEXT_ALLOWED_USER_IDS`; other users receive a permission denial response. The command is intended for cases where the selected duty person is unavailable. It clears today's `last_duty_date` from the current duty record, assigns today's date to the next duty record, notifies the new duty person, and sends an updated mention to the support chat.

Duty selection and reassignment are serialized in PostgreSQL, so concurrent commands cannot assign two people to the same date. Mentions use the native format for the configured transport.

## Integration Tests

The PostgreSQL integration test creates and removes an isolated temporary schema:

```bash
TEST_CONNECTION_STR='postgres://username:password@localhost:5432/watch_bot_test?sslmode=disable' \
  make integration-test
```
