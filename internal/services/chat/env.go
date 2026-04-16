package chat

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// EnvInfo describes the host environment for the running service.
type EnvInfo struct {
	OS             string `json:"os"`
	Arch           string `json:"arch"`
	Hostname       string `json:"hostname"`
	Version        string `json:"version"`
	Shell          string `json:"shell"`
	Timezone       string `json:"timezone"`
	TimezoneOffset string `json:"timezone_offset"`
}

var (
	staticEnvInfo     EnvInfo
	staticEnvInfoOnce sync.Once
)

// CollectEnvInfo gathers host environment info (cached once per process).
func CollectEnvInfo() *EnvInfo {
	staticEnvInfoOnce.Do(func() {
		staticEnvInfo = EnvInfo{
			OS:      runtime.GOOS,
			Arch:    runtime.GOARCH,
			Version: detectOSVersion(),
			Shell:   detectShell(),
		}
		if hostname, err := os.Hostname(); err == nil {
			staticEnvInfo.Hostname = hostname
		}
	})

	now := time.Now()
	info := &EnvInfo{
		OS:             staticEnvInfo.OS,
		Arch:           staticEnvInfo.Arch,
		Hostname:       staticEnvInfo.Hostname,
		Version:        staticEnvInfo.Version,
		Shell:          staticEnvInfo.Shell,
		Timezone:       now.Location().String(),
		TimezoneOffset: now.Format("-07:00"),
	}

	return info
}

// FormatForPrompt formats env info as text for system prompts.
func (e *EnvInfo) FormatForPrompt() string {
	var b strings.Builder
	if e.Timezone != "" {
		if e.TimezoneOffset != "" {
			b.WriteString(fmt.Sprintf("- Timezone: %s (UTC%s)\n", e.Timezone, e.TimezoneOffset))
		} else {
			b.WriteString(fmt.Sprintf("- Timezone: %s\n", e.Timezone))
		}
	}
	b.WriteString(fmt.Sprintf("- OS: %s\n", e.OS))
	b.WriteString(fmt.Sprintf("- Architecture: %s\n", e.Arch))
	if e.Version != "" {
		b.WriteString(fmt.Sprintf("- OS version: %s\n", e.Version))
	}
	if e.Hostname != "" {
		b.WriteString(fmt.Sprintf("- Hostname: %s\n", e.Hostname))
	}
	if e.Shell != "" {
		b.WriteString(fmt.Sprintf("- Default shell: %s\n", e.Shell))
	}
	return b.String()
}

func detectOSVersion() string {
	switch runtime.GOOS {
	case "windows":
		out, err := exec.Command("cmd", "/C", "ver").Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	case "darwin":
		out, err := exec.Command("sw_vers", "-productVersion").Output()
		if err == nil {
			return "macOS " + strings.TrimSpace(string(out))
		}
	default:
		// Linux and other Unix-like systems
		if data, err := os.ReadFile("/etc/os-release"); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(line, "PRETTY_NAME=") {
					val := strings.TrimPrefix(line, "PRETTY_NAME=")
					val = strings.Trim(val, "\"")
					return val
				}
			}
		}
		out, err := exec.Command("uname", "-r").Output()
		if err == nil {
			return "Linux " + strings.TrimSpace(string(out))
		}
	}
	return ""
}

func detectShell() string {
	switch runtime.GOOS {
	case "windows":
		if comspec := os.Getenv("COMSPEC"); comspec != "" {
			return comspec
		}
		return "cmd.exe"
	default:
		if shell := os.Getenv("SHELL"); shell != "" {
			return shell
		}
		return "/bin/sh"
	}
}
