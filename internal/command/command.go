package command

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}

type ServiceController interface {
	Install() error
	Start() error
	Stop() error
	Restart() error
	Status() (string, error)
	Uninstall() error
	Run() error
}

type Options struct {
	Args      []string
	Stdout    io.Writer
	RunCLI    func() error
	RunServer func() error
	Update    func(args []string, stdout io.Writer) error
	Service   ServiceController
	Version   VersionInfo
}

func Execute(opts Options) error {
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}

	args := opts.Args
	if len(args) == 0 {
		return call("cli", opts.RunCLI)
	}

	switch args[0] {
	case "server":
		return call("server", opts.RunServer)
	case "service":
		return executeService(args[1:], stdout, opts.Service)
	case "update":
		return executeUpdate(args[1:], stdout, opts.Update)
	case "version":
		printVersion(stdout, opts.Version)
		return nil
	case "help", "--help", "-h":
		printHelp(stdout)
		return nil
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], HelpText())
	}
}

func executeUpdate(args []string, stdout io.Writer, update func([]string, io.Writer) error) error {
	if update == nil {
		return fmt.Errorf("update controller is not configured")
	}
	return update(args, stdout)
}

func executeService(args []string, stdout io.Writer, svc ServiceController) error {
	if svc == nil {
		return fmt.Errorf("service controller is not configured")
	}
	if len(args) == 0 {
		return fmt.Errorf("missing service action\n\n%s", HelpText())
	}

	switch args[0] {
	case "run":
		return svc.Run()
	case "install":
		return svc.Install()
	case "start":
		return svc.Start()
	case "stop":
		return svc.Stop()
	case "restart":
		return svc.Restart()
	case "status":
		status, err := svc.Status()
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(stdout, status)
		return nil
	case "uninstall":
		return svc.Uninstall()
	default:
		return fmt.Errorf("unknown service action %q\n\n%s", args[0], HelpText())
	}
}

func call(name string, fn func() error) error {
	if fn == nil {
		return fmt.Errorf("%s runner is not configured", name)
	}
	return fn()
}

func printVersion(w io.Writer, v VersionInfo) {
	version := strings.TrimSpace(v.Version)
	if version == "" {
		version = "dev"
	}
	commit := strings.TrimSpace(v.Commit)
	if commit == "" {
		commit = "unknown"
	}
	date := strings.TrimSpace(v.Date)
	if date == "" {
		date = "unknown"
	}
	_, _ = fmt.Fprintf(w, "slimebot %s\ncommit: %s\nbuilt: %s\n", version, commit, date)
}

func printHelp(w io.Writer) {
	_, _ = fmt.Fprint(w, HelpText())
}

func HelpText() string {
	return `Usage:
  slimebot                         Start the CLI TUI
  slimebot server                  Start the web service in the foreground
  slimebot service install         Install the web service
  slimebot service start           Start the web service
  slimebot service stop            Stop the web service
  slimebot service restart         Restart the web service
  slimebot service status          Show web service status
  slimebot service uninstall       Uninstall the web service
  slimebot update --check          Check for updates
  slimebot update                  Update to the latest release
  slimebot update --version vX.Y.Z  Update to a specific release
  slimebot version                 Show version information
  slimebot help                    Show this help
`
}
