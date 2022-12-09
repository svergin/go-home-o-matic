package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/halimath/kvlog"
	"nortal.com/learning-go/skeleton/internal/config"
	"nortal.com/learning-go/skeleton/internal/health"
)

// The main entry point for the service. It wires dependencies and starts the HTTP server.
func main() {
	ctx := context.Background()

	// Perform the wiring by calling providers in valid order.
	cfg := config.Provide(ctx)
	healthHandler := health.Provide()

	// Create the root HTTP multiplexer
	mux := http.NewServeMux()
	// Register health endpoints
	mux.Handle("/health/", http.StripPrefix("/health", healthHandler))

	// Create the server
	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: kvlog.Middleware(kvlog.L, mux),
	}

	kvlog.L.Logs("starting")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Start the server listening for connections. Any error returned
		// that is not http.ErrServerClosed causes the application to fail
		// to start. This usually means that the server cannot bind to the
		// given address.
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			kvlog.L.Logs("http listen failed", kvlog.WithErr(err))
			os.Exit(1)
		}
	}()

	// SIGINT or SIGTERM will terminate the app. Register signal handlers...
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	kvlog.L.Logs("running")

	// ... and wait for a signal to arrive.
	<-c

	kvlog.L.Logs("shutting down")

	// Gracefully shutdown the server
	if err := srv.Close(); err != nil {
		kvlog.L.Logs("http close failed", kvlog.WithErr(err))
	}

	wg.Wait()

	kvlog.L.Logs("shutdown complete")
}
