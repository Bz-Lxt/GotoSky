package logger

import (
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	mu   sync.RWMutex
	base *slog.Logger
)

func Init(level string) *slog.Logger {
	var lv slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lv = slog.LevelDebug
	case "warn", "warning":
		lv = slog.LevelWarn
	case "error":
		lv = slog.LevelError
	default:
		lv = slog.LevelInfo
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lv})
	l := slog.New(h)
	mu.Lock()
	base = l
	mu.Unlock()
	slog.SetDefault(l)
	return l
}

func L() *slog.Logger {
	mu.RLock()
	l := base
	mu.RUnlock()
	if l == nil {
		return slog.Default()
	}
	return l
}
