package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/flyx-ai/heimdall/router"
)

func startServer(ctx context.Context, h http.Handler) error {
	srv := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: h,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	srvErrors := make(chan error, 1)

	go func() {
		slog.InfoContext(ctx, "api server started", "port", "8080")
		srvErrors <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)

	select {
	case err := <-srvErrors:
		slog.ErrorContext(ctx, "server error", "error", err)
		return err
	case sig := <-shutdown:
		ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		slog.InfoContext(ctx, "server shutdown initiated", "cause", sig)

		if err := srv.Shutdown(ctxTimeout); err != nil {
			slog.ErrorContext(ctx, "server shutdown failed", "error", err)
		}

		slog.InfoContext(ctx, "server shutdown completed")
	}

	return nil
}

func setup(ctx context.Context) error {
	return startServer(ctx, router.NewRouter(ctx))
}

func main() {
	ctx := context.Background()
	if err := setup(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
