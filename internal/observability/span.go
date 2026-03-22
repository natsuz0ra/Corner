package observability

import (
	"log/slog"
	"time"
)

func Span(name string, start time.Time) {
	slog.Info("perf_span", "name", name, "ms", time.Since(start).Milliseconds())
}

func MsSince(start time.Time) int64 {
	return time.Since(start).Milliseconds()
}
