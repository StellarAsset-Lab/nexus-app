// Command indexer is the entry point for the Nexus Stellar event indexer.
package main

import (
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "indexer")
	logger.Info("nexus indexer scaffold initialized")
}
