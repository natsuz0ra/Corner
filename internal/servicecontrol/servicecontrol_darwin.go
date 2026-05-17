//go:build darwin

package servicecontrol

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	kservice "github.com/kardianos/service"
)

const (
	darwinServiceLabel      = "com.natsuzora.slimebot"
	darwinLegacyLabel       = serviceName
	darwinLaunchAgentSubdir = "Library/LaunchAgents"
	darwinDefaultPath       = "/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin:/usr/sbin:/sbin"
)

type commandRunner func(string, ...string) (string, error)

type darwinLaunchAgentOptions struct {
	Executable string
	HomeDir    string
	UserName   string
	UID        string
	Program    *program
	RunCommand commandRunner
}

type darwinLaunchAgentService struct {
	executable string
	homeDir    string
	userName   string
	uid        string
	program    *program
	runCommand commandRunner
}

func newServiceBackend(prg *program, cfg *kservice.Config) (serviceBackend, error) {
	opts, err := defaultDarwinLaunchAgentOptions(prg, cfg)
	if err != nil {
		return nil, err
	}
	return newDarwinLaunchAgentService(opts), nil
}

func defaultDarwinLaunchAgentOptions(prg *program, cfg *kservice.Config) (darwinLaunchAgentOptions, error) {
	current, err := user.Current()
	if err != nil {
		return darwinLaunchAgentOptions{}, fmt.Errorf("read current user failed: %w", err)
	}

	homeDir := strings.TrimSpace(current.HomeDir)
	if homeDir == "" {
		homeDir, err = os.UserHomeDir()
		if err != nil || strings.TrimSpace(homeDir) == "" {
			return darwinLaunchAgentOptions{}, errors.New("user home directory not found")
		}
	}

	executable := strings.TrimSpace(cfg.Executable)
	if executable == "" {
		executable = executablePath()
	}

	return darwinLaunchAgentOptions{
		Executable: executable,
		HomeDir:    homeDir,
		UserName:   current.Username,
		UID:        current.Uid,
		Program:    prg,
		RunCommand: runCommand,
	}, nil
}

func newDarwinLaunchAgentService(opts darwinLaunchAgentOptions) *darwinLaunchAgentService {
	run := opts.RunCommand
	if run == nil {
		run = runCommand
	}
	return &darwinLaunchAgentService{
		executable: opts.Executable,
		homeDir:    opts.HomeDir,
		userName:   opts.UserName,
		uid:        opts.UID,
		program:    opts.Program,
		runCommand: run,
	}
}

func (s *darwinLaunchAgentService) Install() error {
	if strings.TrimSpace(s.executable) == "" {
		return errors.New("service executable is not configured")
	}
	if err := os.MkdirAll(filepath.Dir(s.plistPath()), 0o700); err != nil {
		return fmt.Errorf("create LaunchAgents dir failed: %w", err)
	}
	if err := os.MkdirAll(s.logDir(), 0o700); err != nil {
		return fmt.Errorf("create service log dir failed: %w", err)
	}
	if _, err := os.Stat(s.plistPath()); err == nil {
		return fmt.Errorf("service already installed: %s", s.plistPath())
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat service plist failed: %w", err)
	}

	content, err := s.plistContent()
	if err != nil {
		return err
	}
	if err := os.WriteFile(s.plistPath(), []byte(content), 0o644); err != nil {
		return fmt.Errorf("write service plist failed: %w", err)
	}
	return nil
}

func (s *darwinLaunchAgentService) Start() error {
	if err := s.legacyLoadedError(); err != nil {
		return err
	}
	if _, err := os.Stat(s.plistPath()); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("service is not installed: %s", s.plistPath())
		}
		return fmt.Errorf("stat service plist failed: %w", err)
	}
	if _, err := s.runCommand("launchctl", "print", s.serviceTarget()); err != nil {
		if _, err := s.runCommand("launchctl", "bootstrap", s.domain(), s.plistPath()); err != nil {
			return fmt.Errorf("launchctl bootstrap failed for %s: %w", s.plistPath(), err)
		}
	}
	if _, err := s.runCommand("launchctl", "kickstart", "-k", s.serviceTarget()); err != nil {
		return fmt.Errorf("launchctl kickstart failed for %s: %w", s.serviceTarget(), err)
	}
	return nil
}

func (s *darwinLaunchAgentService) Stop() error {
	if _, err := s.runCommand("launchctl", "bootout", s.serviceTarget()); err != nil {
		return fmt.Errorf("launchctl bootout failed for %s: %w", s.serviceTarget(), err)
	}
	return nil
}

func (s *darwinLaunchAgentService) Restart() error {
	if err := s.Stop(); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	return s.Start()
}

func (s *darwinLaunchAgentService) Status() (string, error) {
	if err := s.legacyLoadedError(); err != nil {
		return "", err
	}
	out, err := s.runCommand("launchctl", "print", s.serviceTarget())
	if err != nil {
		if _, statErr := os.Stat(s.plistPath()); statErr == nil {
			return "stopped", nil
		}
		return "", fmt.Errorf("service is not installed: %s", s.plistPath())
	}
	if strings.Contains(out, "state = running") || strings.Contains(out, "pid =") || strings.Contains(out, "PID =") {
		return "running", nil
	}
	return "stopped", nil
}

func (s *darwinLaunchAgentService) Uninstall() error {
	_ = s.Stop()
	if err := os.Remove(s.plistPath()); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("service is not installed: %s", s.plistPath())
		}
		return fmt.Errorf("remove service plist failed: %w", err)
	}
	return nil
}

func (s *darwinLaunchAgentService) Run() error {
	if s.program == nil {
		return errors.New("service program is not configured")
	}
	if err := s.program.Start(nil); err != nil {
		return err
	}

	sigChan := make(chan os.Signal, 3)
	signal.Notify(sigChan, syscall.SIGTERM, os.Interrupt)
	<-sigChan
	signal.Stop(sigChan)

	return s.program.Stop(nil)
}

func (s *darwinLaunchAgentService) plistPath() string {
	return filepath.Join(s.homeDir, darwinLaunchAgentSubdir, darwinServiceLabel+".plist")
}

func (s *darwinLaunchAgentService) logDir() string {
	return filepath.Join(s.homeDir, ".slimebot", "log")
}

func (s *darwinLaunchAgentService) domain() string {
	return "gui/" + s.uid
}

func (s *darwinLaunchAgentService) serviceTarget() string {
	return s.domain() + "/" + darwinServiceLabel
}

func (s *darwinLaunchAgentService) legacyTarget() string {
	return s.domain() + "/" + darwinLegacyLabel
}

func (s *darwinLaunchAgentService) plistContent() (string, error) {
	data := struct {
		Label             string
		Executable        string
		HomeDir           string
		UserName          string
		Path              string
		WorkingDirectory  string
		StandardOutPath   string
		StandardErrorPath string
	}{
		Label:             darwinServiceLabel,
		Executable:        s.executable,
		HomeDir:           s.homeDir,
		UserName:          s.userName,
		Path:              darwinDefaultPath,
		WorkingDirectory:  s.homeDir,
		StandardOutPath:   filepath.Join(s.logDir(), "service.out.log"),
		StandardErrorPath: filepath.Join(s.logDir(), "service.err.log"),
	}

	var b bytes.Buffer
	if err := launchAgentTemplate.Execute(&b, data); err != nil {
		return "", fmt.Errorf("render service plist failed: %w", err)
	}
	return b.String(), nil
}

func (s *darwinLaunchAgentService) legacyLoadedError() error {
	out, err := s.runCommand("launchctl", "print", s.legacyTarget())
	if err != nil {
		return nil
	}
	return fmt.Errorf("旧版 slimebot 服务仍在 launchd 中，可能会阻止新版用户服务启动。\n检测输出：%s\n请先清理旧服务：\n  launchctl bootout %s\n如果旧服务来自系统级 LaunchDaemon，请运行：\n  sudo launchctl bootout system /Library/LaunchDaemons/slimebot.plist\n然后重新执行：slimebot service start", strings.TrimSpace(out), s.legacyTarget())
}

func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%s %s failed: %w\n%s", name, strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

var launchAgentTemplate = template.Must(template.New("launch-agent").Parse(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Disabled</key>
	<false/>
	<key>EnvironmentVariables</key>
	<dict>
		<key>HOME</key>
		<string>{{.HomeDir}}</string>
		<key>USER</key>
		<string>{{.UserName}}</string>
		<key>PATH</key>
		<string>{{.Path}}</string>
	</dict>
	<key>KeepAlive</key>
	<true/>
	<key>Label</key>
	<string>{{.Label}}</string>
	<key>ProgramArguments</key>
	<array>
		<string>{{.Executable}}</string>
		<string>service</string>
		<string>run</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>StandardErrorPath</key>
	<string>{{.StandardErrorPath}}</string>
	<key>StandardOutPath</key>
	<string>{{.StandardOutPath}}</string>
	<key>WorkingDirectory</key>
	<string>{{.WorkingDirectory}}</string>
</dict>
</plist>
`))
