# WatchBot

WatchBot is a monitoring tool that checks the liveness of servers and sends notifications via a specified bot (e.g., Telegram). It periodically probes the servers and reports their status.

## Environment Variables

### Database Configuration
- `CONNECTION_STR`: PostgreSQL connection string (format: `postgres://username:password@host:port/dbname?sslmode=disable`)

### Bot Configuration
- `BOT_TOKEN`: Bot token
- `BOT_API_URL`: Bot API URL
- `MAIN_CHAT_ID`: Main chat ID for notifications
- `BOT_TYPE`: Type of bot to use (can be `telegram` or `vk`)

### Probe Configuration
- `PROBE_DELAY`: Delay between probes in seconds (default: 5)
- `DEAD_PROBE_DELAY`: Delay between probes when server is dead in seconds (default: 60)
- `DEAD_PROBE_THRESHOLD`: Number of dead probes before sending a message (default: 10)
- `DEAD_PROBE_PAUSE`: Pause in minutes before continuing to probe after server is dead (default: 30)
- `UNHEALTHY_THRESHOLD`: Number of unhealthy probes before sending a message (default: 3)
- `UNHEALTHY_DELAY`: Delay between unhealthy probes in seconds (default: 2)

### Working Calendar Configuration
- `START_TIME`: Start of working hours (format: "HH:MM", e.g., "09:00")
- `END_TIME`: End of working hours (format: "HH:MM", e.g., "18:00")
- `DAYS_OFF`: Comma-separated list of days off (e.g., "Saturday,Sunday")

### Logging
- `GRAYLOG_ADDR`: Graylog server address (optional)

## SQL Script for Creating Database

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

