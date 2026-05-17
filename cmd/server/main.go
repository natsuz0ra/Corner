package main

import (
	"errors"
	"os"

	"slimebot/internal/app"
	"slimebot/internal/cliapp"
	"slimebot/internal/command"
	"slimebot/internal/logging"
	"slimebot/internal/runtime"
	"slimebot/internal/servicecontrol"

	_ "slimebot/internal/tools"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	if err := command.Execute(command.Options{
		Args:      os.Args[1:],
		RunCLI:    runCLI,
		RunServer: runServer,
		Service:   mustServiceController(),
		Version: command.VersionInfo{
			Version: version,
			Commit:  commit,
			Date:    date,
		},
	}); err != nil {
		var exitErr cliapp.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

func runCLI() error {
	_, cleanupLogs, _ := logging.Init(logging.Options{Mode: logging.ModeCLI})
	defer cleanupLogs()

	if err := runtime.EnsureAndLoadEnv(); err != nil {
		logging.Error("env_bootstrap_failed", "err", err)
		return err
	}

	if err := cliapp.Run(); err != nil {
		logging.Error("cli_failed", "err", err)
		return err
	}
	return nil
}

func runServer() error {
	_, cleanupLogs, _ := logging.Init(logging.Options{Mode: logging.ModeServer})
	defer cleanupLogs()

	if err := runtime.EnsureAndLoadEnv(); err != nil {
		logging.Error("env_bootstrap_failed", "err", err)
		return err
	}

	if err := app.RunFromEnv(); err != nil {
		logging.Error("server_startup_failed", "err", err)
		return err
	}
	return nil
}

func mustServiceController() command.ServiceController {
	controller, err := servicecontrol.NewController()
	if err != nil {
		return failingServiceController{err: err}
	}
	return controller
}

type failingServiceController struct {
	err error
}

func (f failingServiceController) Install() error          { return f.err }
func (f failingServiceController) Start() error            { return f.err }
func (f failingServiceController) Stop() error             { return f.err }
func (f failingServiceController) Restart() error          { return f.err }
func (f failingServiceController) Status() (string, error) { return "", f.err }
func (f failingServiceController) Uninstall() error        { return f.err }
func (f failingServiceController) Run() error              { return f.err }
