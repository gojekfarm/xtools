# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.9.0]

### Added

- [`xprom`](./xprom): Add `xprom` package to standardize Prometheus metrics with OTEL conventions. ([#55](https://github.com/gojekfarm/xtools/pull/55))
- [`riverkfq`](./riverkfq): Add `riverkfq` package to provide async publishing of messages to Kafka. ([#47](https://github.com/gojekfarm/xtools/pull/47))

## [0.8.1]

### Changed

- [`xload`](./xload): Fixed a collision check regression when a struct pointer is decodable and value for the `key` is
  missing. ([#53](https://github.com/gojekfarm/xtools/pull/53))
- [`xload`](./xload): Added `ErrCast` to let caller know which `key` caused the error. ([#52](https://github.com/gojekfarm/xtools/pull/52))
- [`xpod`](./xpod): Add errors to log delegate when health change fails. ([#50](https://github.com/gojekfarm/xtools/pull/50))
- [`xload`](./xload): Added `ErrDecode` to let caller know which `key` caused the error. ([#48](https://github.com/gojekfarm/xtools/pull/48))

## [0.8.0]

### Added

- [`xkafka`](./xkafka): add slog middleware ([#45](https://github.com/gojekfarm/xtools/pull/45))
- [`xkafka`](./xkafka): add retry middleware ([#44](https://github.com/gojekfarm/xtools/pull/44))

### Changed

- [`xload`](./xload): add key collision detection ([#43](https://github.com/gojekfarm/xtools/pull/43))
- [`xkafka`](./xkafka): `xkafka.ErrorHandler` is now a required
  option ([#41](https://github.com/gojekfarm/xtools/pull/41))

## [0.7.0]

### Added

- [`xpod`](./xpod): expose Checker Handlers without managed
  ServeMux ([#39](https://github.com/gojekfarm/xtools/pull/39))
- [`xload`](./xload): add common parsable types ([#37](https://github.com/gojekfarm/xtools/pull/37))

### Changed

- [`xkafka`](./xkafka): refactor {Run,Start,Close} methods ([#38](https://github.com/gojekfarm/xtools/pull/38))
- [`xpod`](./xpod)[**BREAKING**]: rename HealthChecker => Checker ([#39](https://github.com/gojekfarm/xtools/pull/39))

## [0.6.0]

### Added

- [`xload`](./xload) Add [viper](https://github.com/spf13/viper)
  provider. ([#32](https://github.com/gojekfarm/xtools/pull/32))

### Changed

- [`xload`](./xload) Add context to errors to help with debugging. ([#31](https://github.com/gojekfarm/xtools/pull/31))

## [0.5.0]

### Changed

- [`xkafka`](./xkafka) Improved async & sequential consumer implementations.
  See https://github.com/gojekfarm/xtools/pull/26 for more details.
- [`xkafka`](./xkafka) Upgraded `github.com/confluentinc/confluent-kafka-go` to `v2`

## [0.4.1]

### Changed

- [`xload`](./xload) `xload.Load` supports concurrent loading with `xload.Concurrency` option.
- [`xmap`](./xmap) Added `xmap.Values` function to get values from a map.

## [0.4.0]

### Added

- [`xload`](./xload) Added `xload` package which is a struct first data loading library.
- [`xload/providers/yaml`](./xload/providers/yaml) Added `yaml` provider for `xload` package.

### Changed

- [`xmap`](./xmap) Added `xmap.Merge` function to merge two maps.

## [0.3.0]

### Added

- [`xpod`](./xpod) Added `xpod` package which contains utilities that help implement best practices for health checks
  and more, while building go apps for kubernetes pods.
- [`xkafka/middleware`](./xkafka/middleware)
  - Added Prometheus middleware for `Consumer` and `Producer` implementations.
  - Added Logging MiddlewareFunc.

## [0.2.0]

### Added

- [`xkafka`](./xkafka) Added `xkafka` package with `Producer` and `Consumer` implementations that support middleware &
  HTTP-like handlers.

## [0.1.1]

### Changed

- [`generic/slice`](./generic/slice) All functions now accept typed slices as
  input. ([#4](https://github.com/gojekfarm/xtools/pull/4))

## [0.1.0]

### Added

- [`generic`](./generic) package added
- [`xproto`](./xproto) package added

[Unreleased]: https://github.com/gojekfarm/xtools/compare/v0.9.0...HEAD
[0.9.0]: https://github.com/gojekfarm/xtools/releases/tag/v0.9.0
[0.8.1]: https://github.com/gojekfarm/xtools/releases/tag/v0.8.1
[0.8.0]: https://github.com/gojekfarm/xtools/releases/tag/v0.8.0
[0.7.0]: https://github.com/gojekfarm/xtools/releases/tag/v0.7.0
[0.6.0]: https://github.com/gojekfarm/xtools/releases/tag/v0.6.0
[0.5.0]: https://github.com/gojekfarm/xtools/releases/tag/v0.5.0
[0.4.1]: https://github.com/gojekfarm/xtools/releases/tag/v0.4.1
[0.4.0]: https://github.com/gojekfarm/xtools/releases/tag/v0.4.0
[0.3.0]: https://github.com/gojekfarm/xtools/releases/tag/v0.3.0
[0.2.0]: https://github.com/gojekfarm/xtools/releases/tag/v0.2.0
[0.1.1]: https://github.com/gojekfarm/xtools/releases/tag/v0.1.1
[0.1.0]: https://github.com/gojekfarm/xtools/releases/tag/v0.1.0
