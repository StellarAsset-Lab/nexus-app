// Command api is the entry point for the Nexus read API service.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/StellarAsset-Lab/nexus-app/internal/db"
)

func main() {
	migrate := flag.Bool("migrate", false, "run database migrations and exit")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "api")

	if *migrate {
		runMigrations(logger)
		return
	}

	logger.Info("nexus api scaffold initialized")
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
