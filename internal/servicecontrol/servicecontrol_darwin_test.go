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
	if !strings.HasPrefix(plist, `<?xml version="1.0" encoding="UTF-8"?>`) {
		t.Fatalf("plist should start with raw XML declaration:\n%s", plist)
	}
	if strings.Contains(plist, "&lt;?xml") {
		t.Fatalf("plist XML declaration should not be HTML-escaped:\n%s", plist)
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

func TestDarwinRootCommandsReturnSudoGuidance(t *testing.T) {
	home := t.TempDir()
	svc := newDarwinLaunchAgentService(darwinLaunchAgentOptions{
		Executable: "/opt/slimebot/bin/slimebot",
		HomeDir:    home,
		UserName:   "root",
		UID:        "0",
		RunCommand: func(string, ...string) (string, error) {
			t.Fatal("root guard should run before launchctl")
			return "", nil
		},
	})

	for name, run := range map[string]func() error{
		"install":   svc.Install,
		"start":     svc.Start,
		"uninstall": svc.Uninstall,
		"status": func() error {
			_, err := svc.Status()
			return err
		},
	} {
		err := run()
		if err == nil {
			t.Fatalf("%s should reject root user", name)
		}
		for _, want := range []string{"不要使用 sudo", "slimebot service " + name} {
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("%s error missing %q: %v", name, want, err)
			}
		}
	}
}

func TestDarwinInstallValidatesGeneratedPlistAndExecutable(t *testing.T) {
	home := t.TempDir()
	missingExe := filepath.Join(home, "missing-slimebot")
	var svc *darwinLaunchAgentService
	svc = newDarwinLaunchAgentService(darwinLaunchAgentOptions{
		Executable: missingExe,
		HomeDir:    home,
		UserName:   "alice",
		UID:        "501",
		RunCommand: func(name string, args ...string) (string, error) {
			if name == "plutil" && strings.Join(args, " ") == "-lint "+svc.plistPath() {
				return svc.plistPath() + ": OK", nil
			}
			return "", nil
		},
	})

	err := svc.Install()
	if err == nil {
		t.Fatal("Install should fail when executable is missing")
	}
	for _, want := range []string{svc.plistPath(), "executable", missingExe} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("Install error missing %q: %v", want, err)
		}
	}
}

func TestDarwinStartBootstrapFailureIncludesDiagnostics(t *testing.T) {
	home := t.TempDir()
	var svc *darwinLaunchAgentService
	svc = newDarwinLaunchAgentService(darwinLaunchAgentOptions{
		Executable: "/opt/slimebot/bin/slimebot",
		HomeDir:    home,
		UserName:   "alice",
		UID:        "501",
		RunCommand: func(name string, args ...string) (string, error) {
			joined := strings.Join(args, " ")
			switch {
			case name == "launchctl" && joined == "print gui/501/slimebot":
				return "", errLaunchctlNotFound
			case name == "launchctl" && joined == "print gui/501/com.natsuzora.slimebot":
				return "", errLaunchctlNotFound
			case name == "launchctl" && strings.HasPrefix(joined, "bootstrap gui/501 "):
				return "Bootstrap failed: 5: Input/output error", errors.New("exit status 5")
			case name == "plutil" && joined == "-lint "+svc.plistPath():
				return svc.plistPath() + ": OK", nil
			default:
				return "", nil
			}
		},
	})
	if err := os.MkdirAll(filepath.Dir(svc.plistPath()), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(svc.plistPath(), []byte("plist"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := svc.Start()
	if err == nil {
		t.Fatal("Start should fail when bootstrap fails")
	}
	for _, want := range []string{
		"launchctl bootstrap",
		"gui/501",
		svc.plistPath(),
		"/opt/slimebot/bin/slimebot",
		filepath.Join(home, ".slimebot", "log", "service.out.log"),
		filepath.Join(home, ".slimebot", "log", "service.err.log"),
		"plutil",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("Start error missing %q: %v", want, err)
		}
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

func TestDarwinUninstallRemovesPlistWhenJobIsNotLoaded(t *testing.T) {
	home := t.TempDir()
	svc := newDarwinLaunchAgentService(darwinLaunchAgentOptions{
		Executable: "/opt/slimebot/bin/slimebot",
		HomeDir:    home,
		UserName:   "alice",
		UID:        "501",
		RunCommand: func(name string, args ...string) (string, error) {
			if name == "launchctl" && strings.Join(args, " ") == "bootout gui/501/com.natsuzora.slimebot" {
				return "Could not find service", errLaunchctlNotFound
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

	if err := svc.Uninstall(); err != nil {
		t.Fatalf("Uninstall should remove plist when launchctl bootout misses job: %v", err)
	}
	if _, err := os.Stat(svc.plistPath()); !os.IsNotExist(err) {
		t.Fatalf("plist should be removed, stat err = %v", err)
	}
}

var errLaunchctlNotFound = errors.New("not found")
