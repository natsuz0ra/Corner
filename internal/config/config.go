package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"slimebot/internal/runtime"
)

type Config struct {
	ServerPort        string
	DBPath            string
	Frontend          string
	SkillsRoot        string
	HermesSkillsRoots []string

	ChatUploadRoot   string
	JWTSecret        string
	JWTExpireMinutes int

	ContextHistoryRounds int
	DefaultContextSize   int
}

func Load() Config {
	home := runtime.SlimeBotHomeDir()

	return Config{
		ServerPort:           getEnv("SERVER_PORT", "6247"),
		DBPath:               getPathEnv("DB_PATH", filepath.Join(home, "storage", "data.db")),
		Frontend:             getEnv("FRONTEND_ORIGIN", ""),
		SkillsRoot:           getPathEnv("SKILLS_ROOT", filepath.Join(home, "skills")),
		HermesSkillsRoots:    getPathListEnv("HERMES_SKILLS_ROOTS"),
		ChatUploadRoot:       getPathEnv("CHAT_UPLOAD_ROOT", filepath.Join(home, "storage", "chat_uploads")),
		JWTSecret:            getEnv("JWT_SECRET", ""),
		JWTExpireMinutes:     GetIntEnv("JWT_EXPIRE", 15*24*60),
		ContextHistoryRounds: GetIntEnv("CONTEXT_HISTORY_ROUNDS", 20),
		DefaultContextSize:   GetIntEnv("DEFAULT_CONTEXT_SIZE", 1_000_000),
	}
}

func getPathListEnv(key string) []string {
	raw := getEnv(key, "")
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, string(os.PathListSeparator))
	paths := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			paths = append(paths, runtime.ExpandHome(part))
		}
	}
	return paths
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return fallback
	}
	return value
}

func getPathEnv(key, fallback string) string {
	return runtime.ExpandHome(getEnv(key, fallback))
}

func GetIntEnv(key string, fallback int) int {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
