package main

import (
	"log"
	"os"
	"time"

	"log/slog"

	"github.com/lmittmann/tint"
	"github.com/urfave/cli/v2"
)

func main() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	app := cli.NewApp()

	app.Name = "xkafka"

	app.Commands = []*cli.Command{
		{
			Name:   "sequential",
			Action: runSequentialTest,
		},
		{
			Name:   "sequential-manual",
			Action: runSequentialWithManualCommitTest,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
