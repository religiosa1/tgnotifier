# Simple telegram notification service

A lightweight http client for tg, exposing only two methods:
[sendMessage](https://core.telegram.org/bots/api#sendmessage) and
[getMe](https://core.telegram.org/bots/api#getme).

Allows you to send [telegram](https://telegram.org/) text messages to a small
list of predefined users through a CLI, a REST API or as golang library.

It can be used to send various notifications to your personal account, or
a group chat about pipeline errors/completions, some long running process
finished, etc.

It's intended to be simple to deploy and use, without much configuration.
At the same time it requires some actions from users prior to receiving
notifications, so it can't be used to spam random people.

## Initial Telegram Bot setup

To use the service you will need to connect a telegram bot via a bot token,
which will send the notifications. You can create a bot through
telegram's interface called BotFather, as described in their
[tutorial](https://core.telegram.org/bots/tutorial#getting-ready).

Besides that, you need to know telegram id's of users to whom you want to send
the notifications. You can do that via
the [userinfobot](https://t.me/userinfobot) or alternatively by by sending a
message to your bot DMs and accessing the following url:
`https://api.telegram.org/bot<YOUR BOT TOKEN HERE>/getUpdates`

You user id will be in `result[0].message.from.id` field of response (not to be
confused with chat id: `result[0].message.chat.id`)

Before you can receive notifications from the bot, you must initiate the
communication with it. Go to it's page (through the username, provided by the
BotFather) and click on the "start" button. This applies to every user in your
recipients list that want to get the notifications.

## Installation as a standalone app

The easiest way to install is to grab a binary for you platform from the
release section.

Alternatively if you have [Golang](https://go.dev/) available:

```sh
go install 'github.com/religiosa1/tgnotifier/cmd/tgnotifier@latest'
```

Please, refer to the [App Config](#app-config) section of readme, to see what
configuration options are available to you. At the very least, you must
configure BOT_TOKEN and BOT_API_KEY, and BOT_RECIPIENTS.

It is strongly encouraged to do that in a config file only available for you to
read, so you don't expose your tokens and keys through the environment variables
for the whole system.

See instructions on [Docker image](#docker-image) launch below.

## Usage

### As a CLI util

You can use the app to send notification through CLI
from your shell, scripts, cron, etc.

```sh
tgnotifier send "Your message goes here"
# or from stdin
cat message.md | tgnotifier send
# from stdin as HTML
cat message.html | tgnotifier -p HTML
# To a different recipient
tgnotifier send -r 123456789 "Your message goes here"
# To inform you some long-running task is done:
long-running-foo && tgnotifier send "foo is done!" || tgnotifier send "failed!"
# any shell magic you want:
long-running-foo; status=$?; tgnotifier "foo is done with exit status $status"
```

To get the list of available commands run `tgnotifier --help`.

### As a go library

```sh
go get 'github.com/religiosa1/tgnotifier@latest'
```

```go
import "github.com/religiosa1/tgnotifier"

func main() {
  bot, err := tgnotifier.New("YOUR_BOT_TOKEN_FROM_BOTFATHER")
  if err != nil {
    log.Fatal(err)
  }
  recipientsList := []string {"recipientTgId"}
  err := bot.SendMessage("Hello world!", recipientsList)
  if err != nil {
    log.Fatal(err)
  }
}
```

### As a HTTP service

After installing and _[configuring](#app-config) the app_, to run the server:

```sh
tgnotifier
```

On start, application will try to check if the supplied bot token is correct, by
querying the telegram API. If this process fails, the app immediately exits with
the exit status 2.

If the query is successful, then it launches a HTTP server on the port,
specified in the config (6000 by default), listening for the incoming REST
requests.

If HTTP server failed to launch, application exits with the status 1.

There are only two endpoints:

- `POST /` - [to send notification](#to-send-notification)
- `GET /` - [to get healthcheck](#healthcheck-request)

#### To send notification:

```sh
curl -X POST \
  -H "Content-Type: application/json" \
  -H "x-api-key: YOUR_API_KEY" \
  -d '{"message":"Your message"}' \
  http://localhost:6000/
```

`message` is a required field. Its contents can't be longer than 4096 characters.

You can pass optional `parse_mode` value, to modify, how the passed message is
parsed:

```jsonc
{
	"message": "Your message",
	"parse_mode": "MarkdownV2", // OPTIONAL, defaults to MarkdownV2
	"recipients": ["userid1"] // OPTIONAL, defaults to recipients from config
}
```

Supported `parse_mode` values are:

- MarkdownV2 (default)
- HTML
- Markdown

Please note, that the message should conform to the telegram formatting specs,
as described in [docs](https://core.telegram.org/bots/api#formatting-options)
for example all of the following symbols must be escaped with a '\\' character:
`` _*[]()~`>#+-=|{}.! ``

`recipients` field in the request payload allows to override the default recipient
list provided in the config. If default recipients list is not provided in the
config, then `recipients` field in the payload is required, and attempts to
provide a payload without it or with an empty array will result in 400 error.

Empty recipient array in the payload will always lead to 400 error.

#### Healthcheck request

If you want to check if the service is running ok, you can perform a `GET`
request:

```sh
curl -X GET -H "x-api-key: YOUR_API_KEY" http://localhost:6000/
```

It will perform a query to telegram bot API, verifying that the bot token is
still valid, there's no network outages etc.

## Building app from source

To build the application you need [Go](https://go.dev/) version 1.22 or higher.

Using provided [taskfile](https://taskfile.dev/):

```sh
task build
```

Or manually:

```sh
go build github.com/religiosa1/tgnotifier/cmd/tgnotifier
```

Refer to the go docs on cross compilation and stuff.

### Docker image

You can also launch service via the provided docker image.

**Using environment variables:**

```sh
docker run \
  -e BOT_TOKEN=$(pass bot_token) \
  -e BOT_API_KEY=$(pass bot_api_key) \
  -e BOT_RECIPIENTS=123456789,987654321 \
  -p 6000:6000 \
  ghcr.io/religiosa1/tgnotifier:latest
```

Or create an env file and launch with that:

```sh
docker run --env-file .env -p 6000:6000 ghcr.io/religiosa1/tgnotifier:latest
```

Replace the recipient IDs with your actual Telegram user IDs. You can pass multiple recipients as a comma-separated list.

**Using a config file:**

Alternatively, you can mount a configuration file:

```sh
docker run \
  -v /etc/tgnotifier.yml:/config.yml:ro \
  -p 6000:6000 \
  ghcr.io/religiosa1/tgnotifier:latest
```

**Important notes:**

- **Required**: `BOT_TOKEN` (Telegram bot token)
- **Recommended**: `BOT_API_KEY` (for authentication - without it, anyone can send requests)
- **Optional**: `BOT_RECIPIENTS` (if not set, you must provide recipients in each API request payload)
- See [App Config](#app-config) section for all configuration options
- The container runs as a non-root user (UID 65532) for security
- Health checks should be configured in your orchestration platform (Docker Compose, Kubernetes, etc.) using the `GET /` endpoint

### App config

Configuration of the app can be provided as a yaml config file, or through
environment variables.

Please, check the [config file](./config.yml) included in this repo, to see the
available configuration values.

Supported environmental variables are:

- BOT_TOKEN bot token as provided by BotFather
- BOT_RECIPIENTS list of default recipients' telegram Ids, separated by comma
- BOT_API_KEY your API Key (see bellow)
- BOT_LOG_LEVEL verbosity level of logs, possible values are 'debug', 'info', 'warn', and 'error'
- BOT_LOG_TYPE controls the logger output, possible values are "text" and "json"
- BOT_ADDR address on which we're launching the http server, defaults to "localhost:6000"`
- BOT_CONFIG_PATH path to configuration file

Upon launch, the service tries to load configuration in the following priority order:

1. Config file explicitly specified via the `--config` flag:

   ```sh
   tgnotifier --config /path/to/config.yml
   ```

2. Config file path specified in the `BOT_CONFIG_PATH` environment variable

3. User-specific config file (platform-dependent):

   - **Unix/Linux/macOS**: `$XDG_CONFIG_HOME/tgnotifier/config.yml` (or `~/.config/tgnotifier/config.yml` if `XDG_CONFIG_HOME` is not set)
   - **Windows**: `%APPDATA%\tgnotifier\config.yml`

4. Global config file (platform-dependent):

   - **Unix/Linux/macOS**: `/etc/tgnotifier.yml`
   - **Windows**: `%PROGRAMDATA%\tgnotifier\config.yml`

5. If no config file is found, the service will attempt to load all configuration from environment variables.

This allows you to have a system-wide configuration in the global location, while individual users can override settings with their own user-specific config files.

#### Overriding default config paths with ldflags during the build

You can use ldflags to override the default config file locations at build time.

**Available variables:**

in `github.com/religiosa1/tgnotifier/internal/config`:

- `UserConfigPath` - User-specific config path
- `GlobalConfigPath` - Global/system-wide config path

**For Unix/Linux/macOS:**

Using taskfile's variables:

```sh
# Notice the double slash before the variable to escape variable expansion inside of taskfile
task build USER_CONFIG="\\${XDG_CONFIG_HOME}/tgnotifier/config.yml" GLOBAL_CONFIG="/etc/tgnotifier.yml"
```

Or manually:

```sh
go build \
  -ldflags="-X 'github.com/religiosa1/tgnotifier/internal/config.UserConfigPath=\${XDG_CONFIG_HOME}/tgnotifier/config.yml' \
            -X 'github.com/religiosa1/tgnotifier/internal/config.GlobalConfigPath=/etc/tgnotifier.yml'" \
  github.com/religiosa1/tgnotifier/cmd/tgnotifier
```

**For Windows:**

```sh
go build \
  -ldflags="-X 'github.com/religiosa1/tgnotifier/internal/config.UserConfigPath=\${APPDATA}\\tgnotifier\\config.yml' \
            -X 'github.com/religiosa1/tgnotifier/internal/config.GlobalConfigPath=\${PROGRAMDATA}\\tgnotifier\\config.yml'" \
  github.com/religiosa1/tgnotifier/cmd/tgnotifier
```

**Environment variable expansion:**

The config paths support environment variable expansion at runtime using the `${VAR}` or `$VAR` syntax. Any environment variable can be used as a placeholder:

- **Unix/Linux/macOS examples**: `${XDG_CONFIG_HOME}`, `${HOME}`, `${USER}`, etc.
- **Windows examples**: `${APPDATA}`, `${PROGRAMDATA}`, `${USERPROFILE}`, etc.

### API KEY

You can use API key mechanism, to authorize the incoming request.
It can be any string of characters, but better if it's cryptographically secure.
You can generate an API key with the corresponding command:

```sh
tgnotifier generate-key # it will output a new key to your terminal
# Copy and paste it into the config/put it in env under the BOT_API_KEY
```

This key should be supplied with each http request to the bot via the header
`x-api-key` or with the cookie `X-API-KEY`.

IMPORTANT! Make sure, you don't expose your API key (i.e. don't send those requests
directly from a web page), as anyone who has network access to the service and
has the key can send those notification requests.

If you don't specify an API-key in the config, env, or CLI authorization will
be disabled and service will serve any request.

Please notice, there's no internal rate limiting inside of the service, you
need to handle that, if you want to be safe.

### Example configuration as a service on Linux

Below is an example of setting up the service on Ubuntu Server 22, assuming you
placed the app binary in `/usr/local/bin/tgnotifier`.

Add a user to run the service:

```sh
sudo useradd --system --shell /usr/sbin/nologin tgnotifier
```

Add your configuration in /etc/tgnotifier.yml

Make this configuration readable only to the tgnotifier group:

```sh
chown root:tgnotifier /etc/tgnotifier.yml
chmod 640 /etc/tgnotifier.yml
```

Verify that `tgnotifier` can successfully send notifications:

```sh
# using which here to avoid potential path issues
sudo -u tgnotifier $(which tgnotifier) send -c /etc/tgnotifier.yml "test"
```

Create a systemd init file `/etc/systemd/system/tgnotifier.service`:

```systemd
[Unit]
Description=Telegram Notification Service
After=network-online.target
Wants=network-online.target

[Service]
User=tgnotifier
ExecStart=/usr/local/bin/tgnotifier -c /etc/tgnotifier.yml
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

Enable the service:

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now tgnotifier.service
```

Check status and logs:

```
systemctl status tgnotifier
journalctl -u tgnotifier -f
```

Get the healthcheck to verify it all worked:

```sh
curl -X GET -H "x-api-key: YOUR_API_KEY" http://localhost:6000/
```

By default app launches on localhost only, it will require a reverse proxy to be
accessible from the outside world, such as [caddy](https://caddyserver.com/) or
[nginx](https://nginx.org/).

## License

tgnotifier is MIT licensed.
