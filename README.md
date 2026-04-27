# WatchBot

WatchBot is a duty bot service for Telegram or VK Teams. It exposes health/readiness/metrics endpoints and supports the `\\duty` and `\\next` commands for daily duty rotation.

## Local Development

The repository includes a `Makefile` that pins Go build artifacts to local directories inside the project. This avoids failures in restricted environments where the default Go cache is not writable.

```bash
make test
make build
make run
```

`make build` writes the binary to `.bin/watch_bot`. Temporary Go build artifacts are stored in `.gocache` and `.tmp`, so local runs do not depend on a writable system Go cache.

## Runtime Endpoints

The service starts an HTTP server on port `9000` with the following endpoints:

- `GET /health` - liveness probe
- `GET /ready` - readiness probe
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
- `MAIN_CHAT_ID`: Main chat ID for notifications
- `SUPPORT_CHAT_ID`: Support chat ID for duty notifications and the `\\next` command (required for duty replacement)
- `NEXT_ALLOWED_USER_IDS`: Semicolon-separated list of user IDs allowed to execute `\\next`
- `BOT_TYPE`: Type of bot to use (can be `telegram` or `vk`)
- `RETRY_COUNT`: Number of attempts to send a message (default: 3)
- `RETRY_PAUSE`: Pause between retry attempts in seconds (default: 5)

### Working Calendar Configuration
- `START_TIME`: Start of working hours (format: "HH:MM", e.g., "09:00")
- `END_TIME`: End of working hours (format: "HH:MM", e.g., "18:00")
- `DAYS_OFF`: Comma-separated list of days off (e.g., "Saturday,Sunday")

### Logging
- `GRAYLOG_ADDR`: Graylog server address (optional)

## How To Run

1. Create the database schema:

```sql
create table unusual_days
(
id bigserial constraint unusual_days_pk primary key,
unusual_date date not null
);

create table duties
(
id bigserial constraint duties_pk primary key,
duty_id text not null,
last_duty_date date
);
```

2. Set the required environment variables. Minimal example for Telegram:

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
