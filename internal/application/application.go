package application

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"project_reference/infrastructure/config"
	"project_reference/infrastructure/database"
	"project_reference/infrastructure/rabbit"
	"runtime"
	"syscall"
	"time"
)

func StartApp() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	l = l.With("app_info", slog.GroupValue(
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

	var dbUri string

	if dbUri = viper.GetString("DB_URI"); dbUri == "" {
		l.Error("DB_URI is empty")
		os.Exit(1)
	}

	db, err := database.Init(dbUri)
	if err != nil {
		l.Error("failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			l.Error("failed to close database connection", slog.Any("error", err))
		}
	}()

	var amqpUri string

	if amqpUri = viper.GetString("AMQP_URI"); amqpUri == "" {
		l.Error("AMQP_URI is empty")
		os.Exit(1)
	}

	mq, err := rabbit.NewRabbitMQ(amqpUri)
	if err != nil {
		l.Error("failed to connect to rabbitmq", slog.Any("error", err))
		os.Exit(1)
	}
	defer mq.Close()

	server, err := setupServer(ctx, db, mq)
	if err != nil {
		l.Error("failed to setup server", slog.Any("error", err))
		os.Exit(1)
	}

	serverAddr := fmt.Sprintf("%s:%s", viper.GetString("server.host"), viper.GetString("server.port"))
	srv := &http.Server{
		Addr:         serverAddr,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		Handler:      server,
	}

	l.Info("starting server", slog.String("address", serverAddr))

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Error("listen", "address", serverAddr, "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	l.Info("shutting down server...")

	cancel()
	ctxShutdown, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		l.Error("server shutdown:", slog.Any("error", err))
	}

	<-ctxShutdown.Done()
	l.Info("server exiting")
}
