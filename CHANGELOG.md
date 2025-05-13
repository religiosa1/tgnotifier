# tgnofier changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased
### Added 
- cli: new `send` subcommand to send Telegram messages directly from the command
    line (e.g. for use in cron jobs)
- lib: each bot method now has a `...WithContext` variant to support
    cancellation via context
- lib: support for new fields in the `GetMe` method response:
    `CanConnectToBusiness` and `HasMainWebApp`
- lib: `NewWithClient` method, to supply a custom http.Client for the bot
### Changed
- fix spelling across the project "recepient" -> "recipient"
- lib: `New()` now returns `(*Bot, error)` and may fail on empty token
- lib: `SendMessage` now takes the recipient list as an argument
    instead of being tied to the bot instance to the bot instance
- lib: `tgnotifier` no longer accepts a `slog` logger; instead,
    there's a collection of sentinel errors in the package for granularity
- lib: `GetMeResponse` struct fields names typo fix: 
    `FistName` -> `FirstName`, `CanJoingGroups` -> `CanJoinGroups`
- cli: CLI parsing has been reworked. The `-generate-key` flag is now
    a `generate-key` subcommand (no dash)
- service: response `request_id` is now returned as a ULID instead of a UUID
- service: instead of `env` files implicitely affecting logger output, you 
    can explicitely specify `loggerType` as `text` or `json`
- service: all of the current config values can now be overriden with a cli flag
### Fixed
- parseMode ignored by the service and always resulting in MarkdownV2 messages

## [1.0.0] - 2024.02.26
### Changed
- module renamed to tgnofier, to adhere to golang naming conventions
- config filename env variable renamed to `BOT_CONFIG_FILE`
### Added
- bot itself can now be used as a go module
- LogLevel can now be set in the config/env variables.
- notification endpoint returns 400/500 status depending on internal/user error
- additional checks for non-200 responses on getMe endpoint
### Fixed
- small fixes in logging

## [0.1.0] - 2024.02.25
### Added
- preflight check for BOT token/general connectivity -- attempt to perform
  a request to telgram API, prior to launching the HTTP server
- healthcheck response added at `GET /`
### Changed
- `POST /notify` route moved to the root path `POST /`

## [0.0.1] - 2024.02.19
First public release.