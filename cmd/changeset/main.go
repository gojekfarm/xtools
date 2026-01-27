// Package main provides the changeset CLI for managing multi-module Go releases.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"

	"github.com/gojekfarm/xtools/cmd/changeset/app"
)

var (
	// Build information (set via ldflags).
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

func main() {
	cmd := buildCLI()

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
		syscall.SIGHUP,
	)
	defer cancel()

	if err := cmd.Run(ctx, os.Args); err != nil {
		slog.Error("command failed", "error", err)
		os.Exit(1)
	}
}

func buildCLI() *cli.Command {
	return &cli.Command{
		Name:    "changeset",
		Usage:   "Manage versioning and releases for multi-module Go projects",
		Version: formatVersion(),
		Commands: []*cli.Command{
			buildInitCommand(),
			buildAddCommand(),
			buildStatusCommand(),
			buildVersionCommand(),
			buildPublishCommand(),
			buildTagCommand(),
		},
	}
}

func formatVersion() string {
	return Version + " (commit: " + GitCommit + ", built: " + BuildDate + ")"
}

func buildInitCommand() *cli.Command {
	return &cli.Command{
		Name:        "init",
		Usage:       "Initialize changeset configuration",
		Description: "Creates the .changeset directory with config.json and README.md",
		Action:      app.Init,
	}
}

func buildAddCommand() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Create a new changeset",
		Description: "Interactively create a changeset file documenting your changes.\n\n" +
			"Examples:\n" +
			"  changeset add\n" +
			"  changeset add --empty\n" +
			"  changeset add --open",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "empty",
				Usage: "Create an empty changeset (for changes that don't need releases)",
			},
			&cli.BoolFlag{
				Name:  "open",
				Usage: "Open the created changeset in your editor",
			},
		},
		Action: app.Add,
	}
}

func buildStatusCommand() *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "Show pending changesets and version bumps",
		Description: "Display information about current changesets and computed releases.\n\n" +
			"Examples:\n" +
			"  changeset status\n" +
			"  changeset status --verbose\n" +
			"  changeset status --since=main",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Show full changeset contents and release plan",
			},
			&cli.StringFlag{
				Name:  "output",
				Usage: "Write JSON output to file for CI tools",
			},
			&cli.StringFlag{
				Name:  "since",
				Usage: "Only show changesets since branch/tag",
			},
		},
		Action: app.Status,
	}
}

func buildVersionCommand() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Update versions and changelogs from changesets",
		Description: "Consumes all changesets, updates go.mod files, generates changelogs,\n" +
			"and writes a release manifest for the publish step.\n\n" +
			"Examples:\n" +
			"  changeset version\n" +
			"  changeset version --ignore=examples/xkafka",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  "ignore",
				Usage: "Skip specific modules from versioning",
			},
			&cli.BoolFlag{
				Name:  "snapshot",
				Usage: "Create snapshot versions for testing",
			},
		},
		Action: app.Version,
	}
}

func buildPublishCommand() *cli.Command {
	return &cli.Command{
		Name:  "publish",
		Usage: "Create and push git tags",
		Description: "Reads the release manifest and creates git tags for each module,\n" +
			"then pushes them to the remote.\n\n" +
			"Examples:\n" +
			"  changeset publish\n" +
			"  changeset publish --no-push",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "no-push",
				Usage: "Create tags locally without pushing",
			},
		},
		Action: app.Publish,
	}
}

func buildTagCommand() *cli.Command {
	return &cli.Command{
		Name:        "tag",
		Usage:       "Create git tags without pushing",
		Description: "Creates git tags from the release manifest without pushing.\nUse 'changeset publish' to also push tags.",
		Action:      app.Tag,
	}
}
