# WatchBot

WatchBot is a monitoring tool that checks the liveness of servers and sends notifications via a specified bot (e.g., Telegram). It periodically probes the servers and reports their status.

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
- stops scheduling new probes
- interrupts watchdog wait loops and bot retry pauses
- marks readiness as failed
- gracefully stops the HTTP server with a 10-second timeout

## Environment Variables

### Database Configuration
- `CONNECTION_STR`: PostgreSQL connection string (format: `postgres://username:password@host:port/dbname?sslmode=disable`)

### Bot Configuration
- `BOT_TOKEN`: Bot token
- `BOT_API_URL`: Bot API URL
- `MAIN_CHAT_ID`: Main chat ID for notifications
- `SUPPORT_CHAT_ID`: Support chat ID for duty notifications (required for duty command)
- `BOT_TYPE`: Type of bot to use (can be `telegram` or `vk`)
- `RETRY_COUNT`: Number of attempts to send a message (default: 3)
- `RETRY_PAUSE`: Pause between retry attempts in seconds (default: 5)

### Probe Configuration
- `PROBE_DELAY`: Delay between probes in seconds (default: 5)
- `DEAD_PROBE_DELAY`: Delay between probes when server is dead in seconds (default: 60)
- `DEAD_PROBE_THRESHOLD`: Number of dead probes before sending a message (default: 10)
- `DEAD_PROBE_PAUSE`: Pause in minutes before continuing to probe after server is dead (default: 30)
- `UNHEALTHY_THRESHOLD`: Number of unhealthy probes before sending a message (default: 3)
- `UNHEALTHY_DELAY`: Delay between unhealthy probes in seconds (default: 2)
- `PROBE_TIMEOUT`: Timeout for probe in seconds (default: 3)

### Working Calendar Configuration
- `START_TIME`: Start of working hours (format: "HH:MM", e.g., "09:00")
- `END_TIME`: End of working hours (format: "HH:MM", e.g., "18:00")
- `DAYS_OFF`: Comma-separated list of days off (e.g., "Saturday,Sunday")

### Logging
- `GRAYLOG_ADDR`: Graylog server address (optional)

## How To Run

1. Create the database schema:

```sql
create table servers
(
id bigserial constraint servers_pk primary key,
name text not null,
url  text not null
);

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

export PROBE_DELAY='5'
export DEAD_PROBE_DELAY='60'
export DEAD_PROBE_THRESHOLD='10'
export DEAD_PROBE_PAUSE='30'
export UNHEALTHY_THRESHOLD='3'
export UNHEALTHY_DELAY='2'
export PROBE_TIMEOUT='3'

export RETRY_COUNT='3'
export RETRY_PAUSE='5'
```

3. Add at least one server to monitor:

```sql
insert into servers (name, url) values ('example', 'https://example.com/health');
```

4. Start the service:

```bash
make run
```

5. Verify that the process is up:

```bash
curl http://localhost:9000/health
curl http://localhost:9000/ready
```

## Bot Commands

`\\duty` shows the current duty person. When called, the bot returns a message indicating help is on the way, notifies the person on duty, and on the first assignment of the day also sends a notification to the support chat. For VK Teams, that support notification is sent with HTML parse mode.
