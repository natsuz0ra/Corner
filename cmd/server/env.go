package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"slimebot/internal/runtime"

	"github.com/joho/godotenv"
)

func EnsureAndLoadEnv() error {
	envPath := filepath.Join(runtime.SlimeBotHomeDir(), ".env")
	return ensureAndLoadEnv(envPath, runtime.EnvTemplate())
}

func ensureAndLoadEnv(envPath string, template string) error {
	if strings.TrimSpace(template) == "" {
		return fmt.Errorf("embedded env template is empty")
	}

	if err := os.MkdirAll(filepath.Dir(envPath), os.ModePerm); err != nil {
		return fmt.Errorf("create env dir failed: %w", err)
	}

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		if err := os.WriteFile(envPath, []byte(template), 0o644); err != nil {
			return fmt.Errorf("create env file failed: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("stat env file failed: %w", err)
	} else {
		if err := appendMissingEnvKeys(envPath, template); err != nil {
			return err
		}
	}

	if err := godotenv.Load(envPath); err != nil {
		return fmt.Errorf("load env failed: %w", err)
	}
	return nil
}

func appendMissingEnvKeys(envPath string, template string) error {
	raw, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("read env file failed: %w", err)
	}

	existing := collectEnvKeys(string(raw))
	toAppend := collectMissingTemplateLines(template, existing)
	if len(toAppend) == 0 {
		return nil
	}

	var b strings.Builder
	current := string(raw)
	if current != "" && !strings.HasSuffix(current, "\n") {
		b.WriteString("\n")
	}
	b.WriteString("\n# Auto-appended missing keys\n")
	for _, line := range toAppend {
		b.WriteString(line)
		b.WriteString("\n")
	}

	f, err := os.OpenFile(envPath, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("open env file failed: %w", err)
	}
	defer f.Close()
	if _, err := f.WriteString(b.String()); err != nil {
		return fmt.Errorf("append env file failed: %w", err)
	}
	return nil
}

func collectMissingTemplateLines(template string, existing map[string]struct{}) []string {
	lines := strings.Split(template, "\n")
	missing := make([]string, 0, len(lines))
	seen := make(map[string]struct{})
	for _, line := range lines {
		key, ok := parseEnvKey(line)
		if !ok {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		if _, ok := existing[key]; !ok {
			missing = append(missing, line)
		}
	}
	return missing
}

func collectEnvKeys(content string) map[string]struct{} {
	keys := make(map[string]struct{})
	for _, line := range strings.Split(content, "\n") {
		if key, ok := parseEnvKey(line); ok {
			keys[key] = struct{}{}
		}
	}
	return keys
}

func parseEnvKey(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return "", false
	}
	if strings.HasPrefix(trimmed, "export ") {
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "export "))
	}
	idx := strings.Index(trimmed, "=")
	if idx <= 0 {
		return "", false
	}
	key := strings.TrimSpace(trimmed[:idx])
	if key == "" || strings.Contains(key, " ") {
		return "", false
	}
	return key, true
}
