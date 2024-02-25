# simple-tg-notifier changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased
### Added
- LogLevel can now be set in the config/env variables.
### Fixed
- small fixes in context errors handling and logging

## [1.0.0] - 2024.02.25
### Added
- preflight check for BOT token/general connectivity -- attempt to perform a request 
  to telgram API, prior to launching the HTTP server
- healthcheck response added at `GET /`
### Changed
- `POST /notify` route moved to the root path `POST /`

## [0.0.1] - 2024.02.19
First public release.