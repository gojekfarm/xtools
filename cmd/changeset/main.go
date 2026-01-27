// Package main provides the changeset CLI for managing multi-module Go releases.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gojekfarm/xtools/cmd/changeset/app"
)

var (
	// Build information (set via ldflags).
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

func main() {
	cmd := app.BuildCLI(formatVersion())

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

func formatVersion() string {
	return Version + " (commit: " + GitCommit + ", built: " + BuildDate + ")"
}
