# Simple telegram notifier bot

Bot, that exposes a HTTP Rest API to send messages to the predefined list of users.

## Building
To build the application you need golang version 1.22 or higher.

```sh
go build
```

## Config

Configuration of the app can be provided in as a yaml config file, or through
environment variables.

Please, see the [config file](./config.yml) included in this repo, to see the
available configuration values.

By default, app tries to load the file `./config.yml`. You can override which file
to load calling it with the flag:

```sh
./simple_tg_notifier --config foo.yml
```

Alternatively, you can set all of the configuration with environment variables.

Supported variables are:
- ENV controls the loggger output, possible values are "local", "development", "production"
- BOT_TOKEN bot token as provided by botfather
- BOT_RECEPIENTS list of recepients' telegramIds, separated by comma
- BOT_ADDR address on which we're launching the http server, defaults to "localhost:6000"` 
- BOT_API_KEY your API Key (see bellow)

You need an API key, to authorize the incoming request. It should be a string
of random characters at least 60 characters long. 

You can generate an API key by the application itself, if you launch it with
the corresponding flag

```sh
./simple_tg_notifier --generate-key
```

This key should be supplied with each http request to the bot via the header `x-api-key`
or with the cookie `X-API-KEY`.

## Usage

Run the server:
```sh
./simple_tg_notifier
```

Test that API is working as expected
```sh
curl -X POST \
  -H "Content-Type: application/json" \
  -H "x-api-key: YOUR_API_KEY" \
  -d '{"message":"Your message"}' \
  http://localhost:6000/notify
```

You can pass optional `parse_mode` value, to modify, how the passed message is 
parsed:

```json
{
  "message": "Your message",
  "parse_mode": "MarkdownV2"
}
```

Supported `parse_mode` values are:
- MarkdownV2
- HTML
- Markdown

## License
Simple telegram notifier is MIT licensed.