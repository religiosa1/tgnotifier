# Simple telegram notifier bot

Simple [telegram](https://telegram.org/) notification service, that allows you to 
send text messages to a prefined list of users through a REST API.

It can be used to send you notifications about pipeline errors/completions, 
some long running process finished etc.

It's intended to be simple to deploy and use, without much configuration, notifying 
a small initially known list of users. If you need user permissions, notification 
channels and ability to join/leave them, take a look at 
[notifier](https://github.com/religiosa1/notifier).

## Building

To build the application you need [Go](https://go.dev/) version 1.22 or higher.

```sh
go build
```

Refer to the go docs on crosscompilation and stuff.

## Config

Configuration of the app can be provided as a yaml config file, or through
environment variables.

Please, check the [config file](./config.yml) included in this repo, to see the
available configuration values.

Supported enviromental variables are:
- ENV controls the loggger output, possible values are "local", "development", "production"
- BOT_TOKEN bot token as provided by botfather
- BOT_RECEPIENTS list of recepients' telegramIds, separated by comma
- BOT_ADDR address on which we're launching the http server, defaults to "localhost:6000"` 
- BOT_API_KEY your API Key (see bellow)

By default, app tries to load the file `./config.yml`. You can override which file
to load calling it with the flag:

```sh
./simple_tg_notifier --config foo.yml
```

Alternatively, you can set all of the configuration with environment variables.

You need an API key, to authorize the incoming request. It should be a string
of random characters at least 60 characters long. 

You can generate an API key with the corresponding flag:

```sh
./simple_tg_notifier --generate-key # it will output a new key to your terminal
# Copy and paste it into the config/put it in env under the BOT_API_KEY
```

This key should be supplied with each http request to the bot via the header `x-api-key`
or with the cookie `X-API-KEY`.

## Usage

Run the server:
```sh
./simple_tg_notifier
```
On start, application will try to check if the supplied bot token is correct, by
quering the telegram API. If this process fails, the app immediately exits with
the exit status 2.

If the query is succesfull, then it launches a HTTP server on the port, specified
in the config (6000 by default), listening for the incoming REST requests.

If HTTP server failed to launch, application exits with the status 1.

### To send notification:

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

```json
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

### Healthcheck request

If you want to chec if the service is running ok, you can perform a `GET` request:
```sh
curl -X GET -H "x-api-key: YOUR_API_KEY" http://localhost:6000/
```
It will perform a query to telegram bot API, verifying that the bot token is still
valid, there's no network outages etc.

## License
Simple telegram notifier is MIT licensed.