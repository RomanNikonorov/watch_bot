# WatchBot

WatchBot is a monitoring tool that checks the liveness of servers and sends notifications via a specified bot (e.g., Telegram). It periodically probes the servers and reports their status.

## Run Parameters

- `BOT_TOKEN`: The token for the bot used to send notifications.
- `BOT_API_URL`: The API URL for the bot.
- `MAIN_CHAT_ID`: The main chat ID where notifications will be sent.
- `BOT_TYPE`: The type of bot ("telegram" or "vk").
- `PROBE_DELAY`: The delay in seconds between each probe.
- `DEAD_PROBE_DELAY`: The delay in seconds between probes when the server is dead.
- `DEAD_PROBE_THRESHOLD`: The number of dead probes before sending a message.
- `UNHEALTHY_THRESHOLD`: The number of unhealthy probes before sending a message.
- `UNHEALTHY_DELAY`: The delay in seconds between unhealthy probes.
- `CONNECTION_STR`: The connection string for the database containing the servers to be monitored.

## SQL Script for Creating Database

```sql
create table servers
(
id   bigserial
constraint servers_pk
primary key,
name text not null,
url  text not null
);