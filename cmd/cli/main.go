package main

import (
	"errors"
	"os"

	"slimebot/internal/cliapp"
	"slimebot/internal/logging"
	"slimebot/internal/runtime"

	_ "slimebot/internal/tools"
)

func main() {
	_, cleanupLogs, _ := logging.Init(logging.Options{Mode: logging.ModeCLI})
	defer cleanupLogs()

	if err := runtime.EnsureAndLoadEnv(); err != nil {
		logging.Error("env_bootstrap_failed", "err", err)
		os.Exit(1)
	}

	if err := cliapp.Run(); err != nil {
		var exitErr cliapp.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		logging.Error("cli_failed", "err", err)
		os.Exit(1)
	}
}
