package updater

import (
	"strconv"
	"strings"
)

func NormalizeVersion(version string) string {
	normalized := strings.TrimSpace(version)
	if strings.HasPrefix(normalized, "v") || strings.HasPrefix(normalized, "V") {
		normalized = normalized[1:]
	}
	return normalized
}

func CompareVersions(a, b string) int {
	av, okA := parseSemver(a)
	bv, okB := parseSemver(b)
	if !okA || !okB {
		return 0
	}
	for i := 0; i < len(av); i++ {
		if av[i] > bv[i] {
			return 1
		}
		if av[i] < bv[i] {
			return -1
		}
	}
	return 0
}

func parseSemver(version string) ([3]int, bool) {
	var out [3]int
	normalized := NormalizeVersion(version)
	parts := strings.Split(normalized, ".")
	if len(parts) != 3 {
		return out, false
	}
	for i, part := range parts {
		numeric := part
		if idx := strings.IndexAny(numeric, "-+"); idx >= 0 {
			numeric = numeric[:idx]
		}
		value, err := strconv.Atoi(numeric)
		if err != nil || value < 0 {
			return out, false
		}
		out[i] = value
	}
	return out, true
}
