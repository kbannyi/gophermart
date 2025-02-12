package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

var Log *slog.Logger = slog.Default()

func Initialize() {
	// zero-dependency handler that writes tinted (colorized) logs
	handler := tint.NewHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.Kitchen,
	})

	Log = slog.New(handler)
	slog.SetDefault(Log)
}
