# XTools

[![build][build-workflow-badge]][build-workflow]
[![codecov][coverage-badge]][codecov]
[![docs][docs-badge]][pkg-dev]
[![go-report-card][report-badge]][report-card]

## Introduction

XTools is a submodule based repo to host re-usable Golang code.

## Usage

[API reference][api-docs]

```bash
go get github.com/gojekfarm/xtools/xproto
```

### Contributing Guide

Read our [contributing guide](./CONTRIBUTING.md) to learn about our development process, how to propose bugfixes and improvements, and how to build and test your changes to XTools.

### Release Process

This repo uses Golang [`submodules`](https://github.com/golang/go/wiki/Modules#faqs--multi-module-repositories), to make a new release, make sure to follow the release process described in [RELEASING](RELEASING.md) doc exactly.

## License

XTools is [MIT licensed](./LICENSE).

[build-workflow-badge]: https://github.com/gojekfarm/xtools/workflows/build/badge.svg
[build-workflow]: https://github.com/gojekfarm/xtools/actions?query=workflow%3Abuild
[coverage-badge]: https://codecov.io/gh/gojekfarm/xtools/branch/main/graph/badge.svg?token=ZI56DE8HDH
[codecov]: https://codecov.io/gh/gojekfarm/xtools
[docs-badge]: https://pkg.go.dev/badge/github.com/gojekfarm/xtools
[pkg-dev]: https://pkg.go.dev/github.com/gojekfarm/xtools
[report-badge]: https://goreportcard.com/badge/github.com/gojekfarm/xtools
[report-card]: https://goreportcard.com/report/github.com/gojekfarm/xtools
[api-docs]: https://pkg.go.dev/github.com/gojekfarm/xtools
