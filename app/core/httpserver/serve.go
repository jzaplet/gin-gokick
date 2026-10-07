package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"gokick/app/core/logging"
)

const readHeaderTimeout = 5 * time.Second

const readTimeout = 10 * time.Second

const writeTimeout = 10 * time.Second

const idleTimeout = 120 * time.Second

const shutdownTimeout = 5 * time.Second

func Serve(ctx context.Context, addr string, handler http.Handler, logger *slog.Logger) error {
	var listenConfig net.ListenConfig

	listener, err := listenConfig.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	return serveListener(ctx, listener, handler, logger)
}

func serveListener(ctx context.Context, listener net.Listener, handler http.Handler, logger *slog.Logger) error {
	server := newServer(handler, logger)

	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()

	logger.InfoContext(ctx, "listening", slog.String(logging.KeyAddr, listener.Addr().String()))

	select {
	case err := <-served:
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
	}

	logger.InfoContext(ctx, "shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	if err := <-served; errors.Is(err, http.ErrServerClosed) == false {
		return fmt.Errorf("serve: %w", err)
	}

	return nil
}

func newServer(handler http.Handler, logger *slog.Logger) *http.Server {
	return &http.Server{
		Handler: handler,

		ReadHeaderTimeout: readHeaderTimeout,

		ReadTimeout: readTimeout,

		WriteTimeout: writeTimeout,

		IdleTimeout: idleTimeout,

		ErrorLog: logging.ErrorLog(logger),
	}
}
