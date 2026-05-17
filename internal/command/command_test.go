package command

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestExecuteDefaultsToCLI(t *testing.T) {
	var called []string

	err := Execute(Options{
		Args: []string{},
		RunCLI: func() error {
			called = append(called, "cli")
			return nil
		},
		RunServer: func() error {
			called = append(called, "server")
			return nil
		},
		Service: &fakeServiceController{},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if got := strings.Join(called, ","); got != "cli" {
		t.Fatalf("expected default command to run CLI, got %q", got)
	}
}

func TestExecuteRoutesCLIAndServer(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "explicit cli", args: []string{"cli"}, want: "cli"},
		{name: "server", args: []string{"server"}, want: "server"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var called []string
			err := Execute(Options{
				Args: tt.args,
				RunCLI: func() error {
					called = append(called, "cli")
					return nil
				},
				RunServer: func() error {
					called = append(called, "server")
					return nil
				},
				Service: &fakeServiceController{},
			})
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}
			if got := strings.Join(called, ","); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestExecuteRoutesServiceActions(t *testing.T) {
	actions := []string{"install", "start", "stop", "restart", "status", "uninstall"}

	for _, action := range actions {
		t.Run(action, func(t *testing.T) {
			service := &fakeServiceController{}
			var stdout bytes.Buffer
			err := Execute(Options{
				Args:      []string{"service", action},
				RunCLI:    func() error { return errors.New("cli should not run") },
				RunServer: func() error { return errors.New("server should not run") },
				Service:   service,
				Stdout:    &stdout,
			})
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}
			if got := strings.Join(service.calls, ","); got != action {
				t.Fatalf("expected service action %q, got %q", action, got)
			}
			if action == "status" && !strings.Contains(stdout.String(), "running") {
				t.Fatalf("expected status output, got %q", stdout.String())
			}
		})
	}
}

func TestExecuteVersionAndHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "version", args: []string{"version"}, want: "1.2.3"},
		{name: "help", args: []string{"help"}, want: "slimebot service install"},
		{name: "dash help", args: []string{"--help"}, want: "slimebot service install"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			err := Execute(Options{
				Args:      tt.args,
				RunCLI:    func() error { return errors.New("cli should not run") },
				RunServer: func() error { return errors.New("server should not run") },
				Service:   &fakeServiceController{},
				Stdout:    &stdout,
				Version: VersionInfo{
					Version: "1.2.3",
					Commit:  "abc123",
					Date:    "2026-05-17",
				},
			})
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}
			if !strings.Contains(stdout.String(), tt.want) {
				t.Fatalf("expected output to contain %q, got %q", tt.want, stdout.String())
			}
		})
	}
}

func TestHelpOmitsUpdateYesFlag(t *testing.T) {
	help := HelpText()
	if strings.Contains(help, "update --yes") {
		t.Fatalf("help should not mention update --yes:\n%s", help)
	}
	if !strings.Contains(help, "slimebot update") || !strings.Contains(help, "Update to the latest release") {
		t.Fatalf("help should show bare update command:\n%s", help)
	}
}

func TestExecuteRoutesUpdateCommand(t *testing.T) {
	var calledArgs []string
	var stdout bytes.Buffer
	err := Execute(Options{
		Args:      []string{"update", "--check"},
		RunCLI:    func() error { return errors.New("cli should not run") },
		RunServer: func() error { return errors.New("server should not run") },
		Service:   &fakeServiceController{},
		Stdout:    &stdout,
		Update: func(args []string, w io.Writer) error {
			calledArgs = append(calledArgs, args...)
			_, _ = w.Write([]byte("update ok\n"))
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if got := strings.Join(calledArgs, ","); got != "--check" {
		t.Fatalf("update args = %q, want --check", got)
	}
	if !strings.Contains(stdout.String(), "update ok") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestExecuteRejectsUnknownCommand(t *testing.T) {
	err := Execute(Options{
		Args:      []string{"wat"},
		RunCLI:    func() error { return nil },
		RunServer: func() error { return nil },
		Service:   &fakeServiceController{},
	})
	if err == nil {
		t.Fatal("expected unknown command to fail")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("expected unknown command error, got %v", err)
	}
}

func TestExecuteRejectsUnknownServiceAction(t *testing.T) {
	err := Execute(Options{
		Args:      []string{"service", "bounce"},
		RunCLI:    func() error { return nil },
		RunServer: func() error { return nil },
		Service:   &fakeServiceController{},
	})
	if err == nil {
		t.Fatal("expected unknown service action to fail")
	}
	if !strings.Contains(err.Error(), "unknown service action") {
		t.Fatalf("expected unknown service action error, got %v", err)
	}
}

type fakeServiceController struct {
	calls []string
}

func (f *fakeServiceController) Install() error {
	f.calls = append(f.calls, "install")
	return nil
}

func (f *fakeServiceController) Start() error {
	f.calls = append(f.calls, "start")
	return nil
}

func (f *fakeServiceController) Stop() error {
	f.calls = append(f.calls, "stop")
	return nil
}

func (f *fakeServiceController) Restart() error {
	f.calls = append(f.calls, "restart")
	return nil
}

func (f *fakeServiceController) Status() (string, error) {
	f.calls = append(f.calls, "status")
	return "running", nil
}

func (f *fakeServiceController) Uninstall() error {
	f.calls = append(f.calls, "uninstall")
	return nil
}

func (f *fakeServiceController) Run() error {
	f.calls = append(f.calls, "run")
	return nil
}
