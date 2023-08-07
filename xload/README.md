# xload - Load Anything

`xload` is a struct first data loader. `xload` can be used to load data from environment variables, command line arguments, files, or any remote source. Though `xload` started as a configuration loader, it can be used for business configurations, experiments, internationalization labels, and more.

At high level, `xload` can be used to load data from key-value like sources into structs.

`xload` is inspired by [go-envconfig](https://github.com/sethvargo/go-envconfig). `xload` primarily differs from `go-envconfig` by providing custom data loaders and customizable struct tags.

## Installation

```bash
go get -u github.com/gojekfarm/xtools/xload
```

## Inbuilt Loaders

`xload` comes with a few inbuilt loaders:

- `OSLoader` - Loads data from environment variables
- `PrefixLoader` - Prefixes keys before loading from another loader
- `SerialLoader` - Loads data from multiple loaders, with last non-empty value winning
