package application

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	"github.com/spf13/viper"
	"log/slog"
	"net/http"
	"time"
)

func setupServer(ctx context.Context, infra *Infra) (*http.Server, error) {
	l := slog.Default()

	engine, err := setupDependencies(ctx, infra)
	if err != nil {
		return nil, errors.Wrap(err, "setupServer")
	}

	host := viper.GetString("server.host")
	port := viper.GetString("server.port")
	serverAddr := fmt.Sprintf("%s:%s", host, port)

	srv := &http.Server{
		Addr:         serverAddr,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		Handler:      engine,
	}

	l.Info("starting HTTP server", slog.String("host", host), slog.String("port", port))

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Error("http server error", slog.String("address", serverAddr), slog.Any("error", err))
		}
	}()

	return srv, nil
}
