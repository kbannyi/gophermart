package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger = slog.Default()

func Initialize() {
	var programLevel = new(slog.LevelVar)
	programLevel.Set(slog.LevelDebug)
	Log = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel}))
	slog.SetDefault(Log)
}
