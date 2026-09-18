// Command api is the entry point for the Nexus read API service.
package main

import (
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "api")
	logger.Info("nexus api scaffold initialized")
}
