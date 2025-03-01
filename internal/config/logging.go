package config

import "log/slog"

var slogLevels = map[string]slog.Level{
	"debug":   slog.LevelDebug,
	"info":    slog.LevelInfo,
	"warning": slog.LevelWarn,
	"error":   slog.LevelError,
}

func GetSLogLevel(s string) slog.Level {
	level, ok := slogLevels[s]
	if !ok {
		return slog.LevelInfo
	}

	return level
}
