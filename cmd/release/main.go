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
			{
				Name:  "create",
				Usage: "Create a release by updating all module versions and dependencies",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "version", Usage: "Set the release version explicitly (e.g., v1.2.3)"},
					&cli.BoolFlag{Name: "major", Usage: "Auto-increment the major version"},
					&cli.BoolFlag{Name: "minor", Usage: "Auto-increment the minor version"},
					&cli.BoolFlag{Name: "patch", Usage: "Auto-increment the patch version"},
				},
				Action: CreateRelease,
			},
		},
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
