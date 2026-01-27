package app

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v3"
)

func setupLogger(verbose bool) {
	w := os.Stderr
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	handler := tint.NewHandler(w, &tint.Options{
		Level:      level,
		TimeFormat: time.Kitchen,
		NoColor:    !isatty.IsTerminal(w.Fd()),
	})

	slog.SetDefault(slog.New(handler))
}

// BuildCLI constructs the CLI command tree.
func BuildCLI(version string) *cli.Command {
	var verbose bool

	return &cli.Command{
		Name:    "changeset",
		Usage:   "Manage versioning and releases for multi-module Go projects",
		Version: version,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Usage:       "Enable verbose/debug output",
				Destination: &verbose,
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			setupLogger(verbose)
			return ctx, nil
		},
		Commands: []*cli.Command{
			BuildInitCommand(),
			BuildAddCommand(),
			BuildStatusCommand(),
			BuildVersionCommand(),
			BuildPublishCommand(),
			BuildTagCommand(),
		},
	}
}

// BuildInitCommand creates the init command.
func BuildInitCommand() *cli.Command {
	return &cli.Command{
		Name:        "init",
		Usage:       "Initialize changeset configuration",
		Description: "Creates the .changeset directory with config.json and README.md",
		Action:      Init,
	}
}

// BuildAddCommand creates the add command.
func BuildAddCommand() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Create a new changeset",
		Description: "Interactively create a changeset file documenting your changes.\n\n" +
			"Examples:\n" +
			"  changeset add\n" +
			"  changeset add --empty\n" +
			"  changeset add --open\n" +
			"  changeset add --module libA:minor --module libB:patch --summary \"Add feature\"",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "empty",
				Usage: "Create an empty changeset (for changes that don't need releases)",
			},
			&cli.BoolFlag{
				Name:  "open",
				Usage: "Open the created changeset in your editor",
			},
			&cli.StringSliceFlag{
				Name:  "module",
				Usage: "Module:bump pair for non-interactive mode (e.g., 'libA:minor'). Repeatable.",
			},
			&cli.StringFlag{
				Name:    "summary",
				Aliases: []string{"m"},
				Usage:   "Changeset summary (required with --module)",
			},
		},
		Action: Add,
	}
}

// BuildStatusCommand creates the status command.
func BuildStatusCommand() *cli.Command {
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
		Action: Status,
	}
}

// BuildVersionCommand creates the version command.
func BuildVersionCommand() *cli.Command {
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
		Action: Version,
	}
}

// BuildPublishCommand creates the publish command.
func BuildPublishCommand() *cli.Command {
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
		Action: Publish,
	}
}

// BuildTagCommand creates the tag command.
func BuildTagCommand() *cli.Command {
	return &cli.Command{
		Name:        "tag",
		Usage:       "Create git tags without pushing",
		Description: "Creates git tags from the release manifest without pushing.\nUse 'changeset publish' to also push tags.",
		Action:      Tag,
	}
}
