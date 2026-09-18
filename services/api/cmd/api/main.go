// Command api is the entry point for the Nexus read API service.
package main

import (
	"flag"
	"log/slog"
	"os"
)

func main() {
	migrate := flag.Bool("migrate", false, "run database migrations and exit")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "api")

	if *migrate {
		logger.Error("database migrations are not implemented yet")
		os.Exit(1)
	}

	logger.Info("nexus api scaffold initialized")
}
