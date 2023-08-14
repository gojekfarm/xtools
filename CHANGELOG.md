# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/gojekfarm/xtools/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/gojekfarm/xtools/releases/tag/v0.4.0
[0.3.0]: https://github.com/gojekfarm/xtools/releases/tag/v0.3.0
[0.2.0]: https://github.com/gojekfarm/xtools/releases/tag/v0.2.0
[0.1.1]: https://github.com/gojekfarm/xtools/releases/tag/v0.1.1
[0.1.0]: https://github.com/gojekfarm/xtools/releases/tag/v0.1.0
