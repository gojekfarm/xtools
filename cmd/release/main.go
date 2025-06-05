package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:  "release",
		Usage: "Automate versioning and releasing of multi-module Go projects",
		Commands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "List all modules and their dependencies within the current directory",
				Action: ListModules,
			},
		},
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
