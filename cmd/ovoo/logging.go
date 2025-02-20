package main

import "log/slog"

var slogLevels = map[string]slog.Level{
	"debug":   slog.LevelDebug,
	"info":    slog.LevelInfo,
	"warning": slog.LevelWarn,
	"error":   slog.LevelError,
}

func getSLogLevel(s string) slog.Level {
	level, ok := slogLevels[s]
	if !ok {
		return slog.LevelInfo
	}

	return level
}
