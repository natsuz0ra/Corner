//go:build darwin

package servicecontrol

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDarwinLaunchAgentPlistUsesModernLabelAndUserEnvironment(t *testing.T) {
	home := t.TempDir()
	svc := newDarwinLaunchAgentService(darwinLaunchAgentOptions{
		Executable: "/opt/slimebot/bin/slimebot",
		HomeDir:    home,
		UserName:   "alice",
		UID:        "501",
		RunCommand: func(string, ...string) (string, error) {
			return "", nil
		},
	})

	plist, err := svc.plistContent()
	if err != nil {
		t.Fatalf("plistContent failed: %v", err)
	}

	for _, want := range []string{
		"<string>com.natsuzora.slimebot</string>",
		"<string>/opt/slimebot/bin/slimebot</string>",
		"<string>service</string>",
		"<string>run</string>",
		"<key>EnvironmentVariables</key>",
		"<key>HOME</key>",
		"<string>" + home + "</string>",
		"<key>USER</key>",
		"<string>alice</string>",
		"<key>PATH</key>",
		"<string>/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin:/usr/sbin:/sbin</string>",
		"<key>WorkingDirectory</key>",
		"<string>" + home + "</string>",
		"<key>StandardErrorPath</key>",
		"<string>" + filepath.Join(home, ".slimebot", "log", "service.err.log") + "</string>",
		"<key>StandardOutPath</key>",
		"<string>" + filepath.Join(home, ".slimebot", "log", "service.out.log") + "</string>",
	} {
		if !strings.Contains(plist, want) {
			t.Fatalf("plist missing %q:\n%s", want, plist)
		}
	}
}

func TestDarwinStartUsesBootstrapAndKickstart(t *testing.T) {
	home := t.TempDir()
	var calls []string
	svc := newDarwinLaunchAgentService(darwinLaunchAgentOptions{
		Executable: "/opt/slimebot/bin/slimebot",
		HomeDir:    home,
		UserName:   "alice",
		UID:        "501",
		RunCommand: func(name string, args ...string) (string, error) {
			calls = append(calls, strings.Join(append([]string{name}, args...), " "))
			if name == "launchctl" && len(args) == 2 && args[0] == "print" && args[1] == "gui/501/slimebot" {
				return "", errLaunchctlNotFound
			}
			if name == "launchctl" && len(args) == 2 && args[0] == "print" && args[1] == "gui/501/com.natsuzora.slimebot" {
				return "", errLaunchctlNotFound
			}
			return "", nil
		},
	})
	if err := os.MkdirAll(filepath.Dir(svc.plistPath()), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(svc.plistPath(), []byte("plist"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := svc.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	want := []string{
		"launchctl print gui/501/slimebot",
		"launchctl print gui/501/com.natsuzora.slimebot",
		"launchctl bootstrap gui/501 " + svc.plistPath(),
		"launchctl kickstart -k gui/501/com.natsuzora.slimebot",
	}
	if strings.Join(calls, "\n") != strings.Join(want, "\n") {
		t.Fatalf("launchctl calls:\n%s\nwant:\n%s", strings.Join(calls, "\n"), strings.Join(want, "\n"))
	}
}

func TestDarwinStatusReportsLegacyLoadedJob(t *testing.T) {
	svc := newDarwinLaunchAgentService(darwinLaunchAgentOptions{
		Executable: "/opt/slimebot/bin/slimebot",
		HomeDir:    t.TempDir(),
		UserName:   "alice",
		UID:        "501",
		RunCommand: func(name string, args ...string) (string, error) {
			if name == "launchctl" && strings.Join(args, " ") == "print gui/501/slimebot" {
				return "path = /Library/LaunchDaemons/slimebot.plist", nil
			}
			return "", nil
		},
	})

	_, err := svc.Status()
	if err == nil {
		t.Fatal("expected legacy job error")
	}
	for _, want := range []string{"旧版 slimebot 服务仍在 launchd 中", "sudo launchctl bootout system /Library/LaunchDaemons/slimebot.plist"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error missing %q: %v", want, err)
		}
	}
}

func TestDarwinStartSkipsBootstrapWhenModernJobIsLoaded(t *testing.T) {
	home := t.TempDir()
	var calls []string
	svc := newDarwinLaunchAgentService(darwinLaunchAgentOptions{
		Executable: "/opt/slimebot/bin/slimebot",
		HomeDir:    home,
		UserName:   "alice",
		UID:        "501",
		RunCommand: func(name string, args ...string) (string, error) {
			calls = append(calls, strings.Join(append([]string{name}, args...), " "))
			if name == "launchctl" && strings.Join(args, " ") == "print gui/501/slimebot" {
				return "", errLaunchctlNotFound
			}
			return "state = waiting", nil
		},
	})
	if err := os.MkdirAll(filepath.Dir(svc.plistPath()), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(svc.plistPath(), []byte("plist"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := svc.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	got := strings.Join(calls, "\n")
	if strings.Contains(got, " bootstrap ") {
		t.Fatalf("Start should not bootstrap an already loaded job:\n%s", got)
	}
	if !strings.Contains(got, "launchctl kickstart -k gui/501/com.natsuzora.slimebot") {
		t.Fatalf("Start should kickstart loaded job:\n%s", got)
	}
}

func TestDarwinStopUsesBootoutDomainLabel(t *testing.T) {
	var calls []string
	svc := newDarwinLaunchAgentService(darwinLaunchAgentOptions{
		Executable: "/opt/slimebot/bin/slimebot",
		HomeDir:    t.TempDir(),
		UserName:   "alice",
		UID:        "501",
		RunCommand: func(name string, args ...string) (string, error) {
			calls = append(calls, strings.Join(append([]string{name}, args...), " "))
			return "", nil
		},
	})

	if err := svc.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if got, want := strings.Join(calls, "\n"), "launchctl bootout gui/501/com.natsuzora.slimebot"; got != want {
		t.Fatalf("launchctl call = %q, want %q", got, want)
	}
}

var errLaunchctlNotFound = errors.New("not found")
