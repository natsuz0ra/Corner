package version

import "strings"

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

func Info() BuildInfo {
	return BuildInfo{
		Version: defaultValue(Version, "dev"),
		Commit:  defaultValue(Commit, "unknown"),
		Date:    defaultValue(Date, "unknown"),
	}
}

func defaultValue(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}
