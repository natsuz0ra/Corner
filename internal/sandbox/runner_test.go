package sandbox

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestBuildSeatbeltNetworkOnlyProfileRejectsNetworkWithoutFileRestrictions(t *testing.T) {
	profile := buildSeatbeltNetworkOnlyProfile()

	if !strings.Contains(profile, "(allow default)") {
		t.Fatalf("network-only profile should allow default operations:\n%s", profile)
	}
	if !strings.Contains(profile, "(deny network*)") {
		t.Fatalf("network-only profile should deny network operations:\n%s", profile)
	}
	if strings.Contains(profile, "file-write") || strings.Contains(profile, "subpath") {
		t.Fatalf("network-only profile should not add workspace file restrictions:\n%s", profile)
	}
}

func TestBuildBubblewrapNetworkOnlyArgsUnsharesNetworkWithoutWorkspaceBindings(t *testing.T) {
	req := CommandRequest{CommandName: "curl", CommandArgs: []string{"https://example.com"}, Dir: "/tmp/work"}

	args := buildBubblewrapNetworkOnlyArgs(req)

	if !slices.Contains(args, "--unshare-net") {
		t.Fatalf("network-only args should unshare network: %#v", args)
	}
	if !containsSequence(args, "--bind", "/", "/") {
		t.Fatalf("network-only args should bind the host filesystem read-write: %#v", args)
	}
	if slices.Contains(args, "--ro-bind") {
		t.Fatalf("network-only args should not add read-only workspace restrictions: %#v", args)
	}
	if got := args[len(args)-3:]; !slices.Equal(got, []string{"--", "curl", "https://example.com"}) {
		t.Fatalf("command tail = %#v", got)
	}
}

func containsSequence(values []string, sequence ...string) bool {
	if len(sequence) == 0 || len(sequence) > len(values) {
		return false
	}
	for i := 0; i <= len(values)-len(sequence); i++ {
		if slices.Equal(values[i:i+len(sequence)], sequence) {
			return true
		}
	}
	return false
}

func TestRunCommandDangerFullAccessNetworkDisabledDoesNotUseBareHostCommand(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skipf("network-only sandbox is unsupported on %s", runtime.GOOS)
	}
	policy, err := NewPolicy(Config{
		Mode:    ModeDangerFullAccess,
		CWD:     t.TempDir(),
		Network: NetworkPolicy{Enabled: false},
	})
	if err != nil {
		t.Fatalf("new policy: %v", err)
	}

	origLookPath := lookPath
	origCommandContext := commandContext
	defer func() {
		lookPath = origLookPath
		commandContext = origCommandContext
	}()

	lookPath = func(file string) (string, error) {
		switch {
		case runtime.GOOS == "darwin" && file == "/usr/bin/sandbox-exec":
			return file, nil
		case runtime.GOOS == "linux" && file == "bwrap":
			return "/usr/bin/bwrap", nil
		default:
			return "", errors.New("unexpected executable lookup: " + file)
		}
	}
	commandContext = func(_ context.Context, name string, args ...string) *exec.Cmd {
		t.Helper()
		if name == "echo" {
			t.Fatalf("danger-full-access with disabled network used bare host command")
		}
		if runtime.GOOS == "darwin" && name != "/usr/bin/sandbox-exec" {
			t.Fatalf("darwin command = %q, want sandbox-exec", name)
		}
		if runtime.GOOS == "linux" && name != "/usr/bin/bwrap" {
			t.Fatalf("linux command = %q, want bwrap", name)
		}
		return exec.CommandContext(context.Background(), "true", args...)
	}

	_, err = RunCommand(context.Background(), policy, CommandRequest{
		CommandName: "echo",
		CommandArgs: []string{"ok"},
		Dir:         t.TempDir(),
	})
	if err != nil {
		t.Fatalf("run command: %v", err)
	}
}
