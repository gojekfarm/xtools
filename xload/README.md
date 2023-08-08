# xload - Load Anything

xload is a struct first data loader that simplifies the process of loading data from various sources into Go structs. It is designed to handle diverse use cases, from configuration loading to internationalization labels and even experiments.

## Why xload?

We often encounter the need to load data from various sources, such as environment variables, command-line arguments, files, or remote configurations. This data may include configuration settings, experiment parameters, internationalization labels, design assets and more. Implementing these data loading methods can lead to boilerplate code, making the codebase hard to maintain and less readable.

xload is inspired by the popular go-envconfig library but extends it with custom data loaders and allows customizing struct tags, making it flexible for different use cases.

## Installation

```bash
go get -u github.com/gojekfarm/xtools/xload
```

## How xload Works

At a high level, xload takes a Go struct annotated with tags and populates it with data from any source. The sources can be anything that implements the `Loader` interface. By using xload, you can separate the data loading process from how and where the data is used.

### Loading from Environment Variables

Let's take a look at a simple example of how to use xload to load data from environment variables into a struct:

```go
type AppConfig struct {
    // your application config goes here
}

func DefaultAppConfig() AppConfig {
    return AppConfig{
        // set default values here
    }
}

func main() {
    ctx := context.Background()
    cfg := DefaultAppConfig()

    err := xload.Load(ctx, &cfg)
    if err != nil {
        panic(err)
    }

    // use cfg
}
```

### Custom Types

With xload, you can also use custom types, such as time.Duration, to make the configuration more descriptive and easier to understand:

For example, timeouts need not be ambiguous `int` values, they can be typed `time.Duration` values:

```go
type AppConfig struct {
    Timeout time.Duration `env:"TIMEOUT"` // TIMEOUT=5s will be parsed as 5 seconds
}
```

This helps you do away with naming conventions like `*_TIMEOUT_MS` or `*_TIMEOUT_SECS`.

### Nested Structs

**xload** supports nested structs, slices, maps, and custom types making it easy to group, reuse and maintain complex configurations.

```go
type HTTPConfig struct {
    Port int `env:"PORT"`
    Host string `env:"HOST"`
}

type AppConfig struct {
    Service1 HTTPConfig `env:",prefix=SERVICE1_"`
    Service2 HTTPConfig `env:",prefix=SERVICE2_"`
}
```

## Inbuilt Loaders

`xload` comes with a few inbuilt loaders:

- `OSLoader` - Loads data from environment variables
- `PrefixLoader` - Prefixes keys before loading from another loader
- `SerialLoader` - Loads data from multiple loaders, with last non-empty value winning

## Custom Loaders

You can also write your own custom loaders by implementing the `Loader` interface:

```go
type Loader interface {
    Load(ctx context.Context, key string) (string, error)
}
```
