package cliapp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"

	"slimebot/internal/app"
	"slimebot/internal/logging"
)

type ExitError struct {
	Code int
}

func (e ExitError) Error() string {
	return fmt.Sprintf("cli exited with status %d", e.Code)
}

func Run() error {
	slimeApp, err := app.RunCLIHeadless()
	if err != nil {
		return fmt.Errorf("cli headless start failed: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()
		slimeApp.Close(shutdownCtx)
	}()

	apiURL := fmt.Sprintf("http://%s", slimeApp.Addr())
	cliToken := slimeApp.CLIToken()

	logging.Info("cli_headless_ready", "api_url", apiURL)

	cliEntry := FindCLIEntry()
	if cliEntry == "" {
		return fmt.Errorf("cli/cli.cjs not found. Run 'npm run build:cli' first")
	}

	cmd := exec.Command(FindNode(), cliEntry,
		"--api-url", apiURL,
		"--cli-token", cliToken,
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"SLIMEBOT_API_URL="+apiURL,
		"SLIMEBOT_CLI_TOKEN="+cliToken,
	)

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return ExitError{Code: exitErr.ExitCode()}
		}
		return fmt.Errorf("cli process failed: %w", err)
	}
	return nil
}

func FindCLIEntry() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	base := filepath.Dir(exe)

	candidates := []string{
		filepath.Join(base, "cli", "cli.cjs"),
		filepath.Join(base, "..", "cli", "cli.cjs"),
	}

	wd, _ := os.Getwd()
	candidates = append(candidates,
		filepath.Join(wd, "cli", "cli.cjs"),
	)

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func FindNode() string {
	if nodePath := os.Getenv("NODE_PATH"); nodePath != "" {
		return nodePath
	}

	name := "node"
	if runtime.GOOS == "windows" {
		name = "node.exe"
	}

	if p, err := exec.LookPath(name); err == nil {
		return p
	}

	if runtime.GOOS == "windows" {
		paths := []string{
			`C:\Program Files\nodejs\node.exe`,
			`C:\Program Files (x86)\nodejs\node.exe`,
		}
		if nvmHome := os.Getenv("NVM_HOME"); nvmHome != "" {
			paths = append([]string{filepath.Join(nvmHome, "node.exe")}, paths...)
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}

	return name
}
