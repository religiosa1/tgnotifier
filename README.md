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
At the same time it requires some actions from users prior to recieving 
notifications, so it can't be used to spam random people.

## Initial Telegram Bot setup

To use the service you will need to connect a telegram bot via a bot token, 
which will send the notifications. You can create a bot through
telegram's interface called BotFather, as described in their 
[tutorial](https://core.telegram.org/bots/tutorial#getting-ready).

Besides that, you need to know telegram id's of users to whom you want to 
send the notifications. You can do that via the [userinfobot](https://t.me/userinfobot).

Before you can recieve notifications from the bot, you must iniate the communication
with it. Go to it's page (through the username, provided by the BotFather) and
click on the "start" button. This applies to every user in your recipients list
that want to get the notifications.

## Usage

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

Run the server:
```sh
./tgnotifier
```
On start, application will try to check if the supplied bot token is correct, by
quering the telegram API. If this process fails, the app immediately exits with
the exit status 2.

If the query is succesfull, then it launches a HTTP server on the port, specified
in the config (6000 by default), listening for the incoming REST requests.

If HTTP server failed to launch, application exits with the status 1.

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
  "parse_mode": "MarkdownV2" // OPTIONAL
}
```

Supported `parse_mode` values are:
- MarkdownV2 (default)
- HTML
- Markdown

Please note, that the message should conform to the telegram formatting specs,
as described in the [docs](https://core.telegram.org/bots/api#formatting-options)
for example all of the following symbols must be escaped with a '\\' character: 
```_*[]()~`>#+-=|{}.!```

#### Healthcheck request

If you want to chec if the service is running ok, you can perform a `GET` request:
```sh
curl -X GET -H "x-api-key: YOUR_API_KEY" http://localhost:6000/
```
It will perform a query to telegram bot API, verifying that the bot token is still
valid, there's no network outages etc.

### As a CLI util

You can use the app to send notification through CLI
from your shell, scripts, cron, etc.

```sh
./tgnotifier send "Your message goes here"
# or from stdin
cat message.md | ./tgnotifier send
# from stdin as HTML
cat message.html | ./tgnotifier -p HTML
# To a different recipient
./tgnotifier send "Your message goes here"
```

## Installation as a standalone app
The easiest way to install is to grab a binary for you platform from the
release section. 

Alternatively if you have [Golang](https://go.dev/) available:

```sh
go install 'github.com/religiosa1/tgnotifier/cmd/tgnotifier@latest'
```

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

Refer to the go docs on crosscompilation and stuff.

You can use ldflags, to override the default config file location.

Using taskfile's variables:

```sh
task build DEFAULT_CONFIG=/etc/tgnotifier.yml
```

or manually:

```sh
go build \
  -ldflags="-X 'github.com/religiosa1/tgnotifier/internal/config.DefaultConfigPath=/etc/tgnotifier.yml'" \
   github.com/religiosa1/tgnotifier/cmd/tgnotifier
```


### App config

Configuration of the app can be provided as a yaml config file, or through
environment variables.

Please, check the [config file](./config.yml) included in this repo, to see the
available configuration values.

Supported enviromental variables are:
- BOT_TOKEN bot token as provided by BotFather
- BOT_RECIPIENTS list of recipients' telegram Ids, separated by comma
- BOT_API_KEY your API Key (see bellow)
- BOT_LOG_LEVEL verbosity level of logs, possible values are 'debug', 'info', 'warn', and 'error'
- BOT_LOG_TYPE controls the loggger output, possible values are "text" and "json"
- BOT_ADDR address on which we're launching the http server, defaults to "localhost:6000"`
- BOT_CONFIG_PATH path to configuration file

Upon launch, service tries to load the configuration file specified in the 
`BOT_CONFIG_PATH` environment variable, or `./config.yml` if it's empty. 
You can override which file to load calling it with the flag:

```sh
./tgnotifier --config foo.yml
```

Alternatively, you can set all of the configuration with environment variables.

### API KEY
You can use API key mechanism, to authorize the incoming request. 
It can be any string of characters, but better if it's cryptografically secure.
You can generate an API key with the corresponding command:

```sh
./tgnotifier generate-key # it will output a new key to your terminal
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


## License
tgnotifier is MIT licensed.