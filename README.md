# Simple telegram notifier bot

Bot, that exposes a HTTP Rest API to send messages to the predefined list of users.

## Usage

Run the server:
```sh
./simple_tg_norifier
```

Test that API is working as expected
```sh
curl -X POST \
  -H "Content-Type: application/json" \
  -H "x-api-key: YOUR_API_KEY" \
  -d '{"message":"Your message"}' \
  http://localhost:6000/notify
```