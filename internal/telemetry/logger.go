package telemetry

import (
	"log/slog"
	"os"
	"sync"
)

var (
	once sync.Once
	Log  *slog.Logger
)

// Init initialises the global structured logger.
// It outputs JSON so that log aggregators (ELK, Loki, CloudWatch, etc.) can
// parse entries without extra work.
func Init() {
	once.Do(func() {
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		})
		Log = slog.New(handler)
		slog.SetDefault(Log) // makes stdlib log.Print* also go through slog
	})
}

// Convenience wrappers so callers can write telemetry.Info(...) instead of
// logger.Log.Info(...)

func Info(msg string, args ...any)  { Log.Info(msg, args...) }
func Warn(msg string, args ...any)  { Log.Warn(msg, args...) }
func Error(msg string, args ...any) { Log.Error(msg, args...) }
func Debug(msg string, args ...any) { Log.Debug(msg, args...) }
