package cli_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/gojekfarm/xtools/x/cli"
)

type Config struct {
	Debug bool `flag:"debug" env:"DEBUG" default:"false" global-flag:"true" flag-usage:"enable debug mode"`
	Log   struct {
		Level string `flag:"level" env:"LEVEL" default:"info"`
	} `flag-prefix:"log" env-prefix:"LOG" global-flag:"true"`
	Host     string `flag:"host" env:"HOST" default:"localhost" sub-commands:"server"`
	Port     int    `flag:"port" env:"PORT" default:"8080" sub-commands:"server"`
	Database struct {
		URL  string `flag:"url" env:"URL" default:"mongodb://localhost:27017"`
		Name string `flag:"name" env:"NAME" default:"staging"`
	} `sub-commands:"migrate,server" flag-prefix:"db" env-prefix:"DB"`
}

func ExampleNew() {
	cfg := new(Config)

	c := cli.New(
		"my-command",
		cli.ConfigObject(cfg),
		cli.Commands{
			{
				Name: "migrate",
				Run: func(ctx context.Context, cfg interface{}) error {
					fmt.Println("run migrate")
					if v, ok := cfg.(*Config); !ok {
						return errors.New("not a Config object")
					} else {
						fmt.Printf("%+v\n", v)
					}
					return nil
				},
				Commands: []cli.Command{
					{
						Name: "up",
						Run: func(ctx context.Context, cfg interface{}) error {
							fmt.Println("run migrate up")
							if v, ok := cfg.(*Config); !ok {
								return errors.New("not a Config object")
							} else {
								fmt.Printf("%+v\n", v)
							}
							return nil
						},
					},
				},
			},
			{
				Name: "server",
				Run: func(ctx context.Context, cfg interface{}) error {
					fmt.Println("run server")
					if v, ok := cfg.(*Config); !ok {
						return errors.New("not a Config object")
					} else {
						fmt.Printf("%+v\n", v)
					}
					return nil
				},
			},
		},
	)

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	c.SetArgs([]string{"server", "--help"})
	if err := c.Run(ctx); err != nil {
		os.Exit(1)
	}

	// Output:
	//
}
