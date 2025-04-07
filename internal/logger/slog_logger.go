package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/jmoloko/taskmange/internal/config"
)

type SLogLogger struct {
	logger *slog.Logger
	file   *os.File
}

func NewSLogLogger(cfg config.LoggerConfig) Logger {
	// Set default output to stdout
	var output io.Writer = os.Stdout
	var file *os.File

	if cfg.File != "" {
		var err error
		file, err = os.OpenFile(cfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			slog.Error("Failed to open log file, falling back to stdout only", "error", err)
		} else {
			output = io.MultiWriter(os.Stdout, file)
		}
	}

	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Convert time to RFC3339Nano format
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					return slog.String(a.Key, t.Format(time.RFC3339Nano))
				}
			}
			return a
		},
	}

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	logger := slog.New(handler).With(
		"service", cfg.ServiceName,
		"env", cfg.Environment,
	)

	return &SLogLogger{
		logger: logger,
		file:   file,
	}
}

func (l *SLogLogger) Debug(msg string, args ...interface{}) {
	l.log(context.Background(), slog.LevelDebug, msg, args...)
}

func (l *SLogLogger) Info(msg string, args ...interface{}) {
	l.log(context.Background(), slog.LevelInfo, msg, args...)
}

func (l *SLogLogger) Warn(msg string, args ...interface{}) {
	l.log(context.Background(), slog.LevelWarn, msg, args...)
}

func (l *SLogLogger) Error(msg string, args ...interface{}) {
	l.log(context.Background(), slog.LevelError, msg, args...)
}

func (l *SLogLogger) Fatal(msg string, args ...interface{}) {
	l.log(context.Background(), slog.LevelError, msg, args...)
	os.Exit(1)
}

func (l *SLogLogger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

func (l *SLogLogger) WithFields(fields map[string]interface{}) Logger {
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}

	return &SLogLogger{
		logger: l.logger.With(attrs...),
		file:   l.file,
	}
}

func (l *SLogLogger) log(ctx context.Context, level slog.Level, msg string, args ...interface{}) {
	// If first arg is format string and there are more args, treat as printf
	if len(args) > 0 {
		if format, ok := args[0].(string); ok && len(args) > 1 {
			msg = fmt.Sprintf(format, args[1:]...)
			args = nil
		}
	}

	attrs := argsToAttrs(args)

	l.logger.Log(ctx, level, msg, attrs...)
}

func argsToAttrs(args []interface{}) []any {
	if len(args) == 0 {
		return nil
	}

	attrs := make([]any, 0, len(args))
	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			key = "arg"
		}

		var value interface{} = "<missing>"
		if i+1 < len(args) {
			value = args[i+1]
		}

		attrs = append(attrs, key, value)
	}

	return attrs
}
