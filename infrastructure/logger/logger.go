package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"time"
)

var (
	defaultLogger *slog.Logger
	once          sync.Once
)

type Options struct {
	Writer   io.Writer
	Level    slog.Leveler
	AddTrace bool
	AppName  string
}

type contextKey string

const traceIDKey contextKey = "traceID"

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

func getTraceID(ctx context.Context) string {
	if v := ctx.Value(traceIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}

	return ""
}

func Init(opts Options) {
	once.Do(func() {
		if opts.Writer == nil {
			opts.Writer = os.Stderr
		}
		if opts.Level == nil {
			opts.Level = slog.LevelInfo
		}

		handlerOpts := &slog.HandlerOptions{
			Level: opts.Level,
		}

		h := slog.NewJSONHandler(opts.Writer, handlerOpts)

		l := slog.New(h).With(
			"ts", slog.TimeValue(time.Now()),
			"app", opts.AppName,
			"os", runtime.GOOS,
			"go", runtime.Version(),
		)

		defaultLogger = l
	})
}

func Logger(ctx context.Context) *slog.Logger {
	if defaultLogger == nil {
		Init(Options{})
	}

	if traceID := getTraceID(ctx); traceID != "" {
		return defaultLogger.With("trace_id", traceID)
	}

	return defaultLogger
}

//func handler(ctx context.Context) {
//	ctx = logger.WithTraceID(ctx, "xyz-123")
//	logger.Logger(ctx).Info("user logged in", slog.String("user_id", "42"))
//}
