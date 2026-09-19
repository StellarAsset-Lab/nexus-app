// Command api is the entry point for the Nexus read API service.
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/StellarAsset-Lab/nexus-app/internal/db"
	"github.com/StellarAsset-Lab/nexus-app/internal/httpapi"
	"github.com/StellarAsset-Lab/nexus-app/internal/network"
)

func main() {
	migrate := flag.Bool("migrate", false, "run database migrations and exit")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "api")

	if *migrate {
		runMigrations(logger)
		return
	}

	if err := run(logger); err != nil {
		logger.Error("api exited with error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is not set")
	}

	bindAddr := os.Getenv("API_BIND_ADDR")
	if bindAddr == "" {
		bindAddr = ":8080"
	}

	netConfig, err := network.LoadConfig()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conn, err := db.Open(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	server := httpapi.NewServer(logger, conn, netConfig)

	httpServer := &http.Server{
		Addr:              bindAddr,
		Handler:           server.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("api listening", "addr", bindAddr, "network", netConfig.Name)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	case err := <-serveErr:
		return err
	}
}

func runMigrations(logger *slog.Logger) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		logger.Error("DATABASE_URL is not set")
		os.Exit(1)
	}

	ctx := context.Background()

	conn, err := db.Open(ctx, databaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	if err := db.Migrate(ctx, conn); err != nil {
		logger.Error("migration failed", "error", err)
		os.Exit(1)
	}

	logger.Info("database migrations applied successfully")
}
