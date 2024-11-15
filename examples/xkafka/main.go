package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func main() {
	// Configure zerolog
	zerolog.TimeFieldFormat = time.Kitchen
	console := zerolog.ConsoleWriter{Out: os.Stderr}
	log.Logger = log.Output(console)

	app := cli.NewApp()

	app.Name = "xkafka"

	flags := []cli.Flag{
		&cli.IntFlag{
			Name:    "partitions",
			Aliases: []string{"p"},
			Value:   4,
		},
		&cli.IntFlag{
			Name:    "consumers",
			Aliases: []string{"c"},
			Value:   2,
		},
		&cli.IntFlag{
			Name:    "messages",
			Aliases: []string{"m"},
			Value:   10,
		},
		&cli.IntFlag{
			Name:    "concurrency",
			Aliases: []string{"cc"},
			Value:   1,
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:   "basic",
			Usage:  "Run basic consumer example",
			Flags:  flags,
			Action: runBasic,
		},
		{
			Name:  "batch",
			Usage: "Run batch consumer example",
			Flags: append(flags,
				&cli.IntFlag{
					Name:    "batch-size",
					Aliases: []string{"bs"},
					Value:   10,
				},
			),
			Action: runBatch,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err)
	}
}
