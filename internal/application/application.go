package application

import (
	"context"
	"log/slog"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"project_reference/infrastructure/config"
	"runtime"
	"syscall"
	"time"
)

func StartApp() {
	ctx, cancel := context.WithCancel(context.Background())

	l := slog.New(slog.NewJSONHandler(os.Stderr, nil)).With("app_info", slog.GroupValue(
		slog.String("os", runtime.GOOS),
		slog.String("go_version", runtime.Version()),
		slog.Int("num_cpu", runtime.NumCPU()),
		slog.Int("num_goroutine", runtime.NumGoroutine()),
	))
	slog.SetDefault(l)

	if err := config.LoadConfig(); err != nil {
		l.Error("failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	infra, err := setupInfra()
	if err != nil {
		l.Error("failed to setup infra", slog.Any("error", err))
		os.Exit(1)
	}

	defer infra.Close()

	server, err := setupServer(ctx, infra)
	if err != nil {
		l.Error("failed to setup server", slog.Any("error", err))
		os.Exit(1)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	l.Info("shutting down...")

	cancel()

	ctxShutdown, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(ctxShutdown); err != nil {
		l.Error("server shutdown error", slog.Any("error", err))
	} else {
		l.Info("server gracefully stopped")
	}
}
