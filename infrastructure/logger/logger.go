package logger

import (
	"golang-template/internal/domain/logger"
	"io"
	"log/slog"
	"os"
	"time"
)

type Logger struct {
	base *slog.Logger
	opts *Options
}

type Options struct {
	Writer  io.Writer
	Level   slog.Leveler
	AppName string
}

func Init(opts Options) *Logger {
	if opts.Writer == nil {
		opts.Writer = os.Stderr
	}

	if opts.Level == nil {
		opts.Level = slog.LevelDebug
	}

	handlerOpts := &slog.HandlerOptions{
		Level: opts.Level,
	}

	h := slog.NewJSONHandler(opts.Writer, handlerOpts)

	l := slog.New(h).With(
		slog.Any("ts", slog.TimeValue(time.Now())),
		slog.String("app", opts.AppName),
	)

	return &Logger{
		base: l,
		opts: &opts,
	}
}

func (l *Logger) Info(msg string, args ...any) {
	l.base.Info(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.base.Warn(msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.base.Debug(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.base.Error(msg, args...)
}

func (l *Logger) With(args ...any) logger.Loggerer {
	newLogger := *l
	newLogger.base = l.base.With(args...)
	return &newLogger
}
