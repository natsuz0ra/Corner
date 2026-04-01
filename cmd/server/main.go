package main

import (
	"log/slog"
	"os"

	"slimebot/internal/app"

	_ "slimebot/internal/tools"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	if err := EnsureAndLoadEnv(); err != nil {
		slog.Error("env_bootstrap_failed", "err", err)
		os.Exit(1)
	}

	if err := app.RunFromEnv(); err != nil {
		slog.Error("server_startup_failed", "err", err)
		os.Exit(1)
	}
}
