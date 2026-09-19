// Command worker is the entry point for the Nexus reconciliation worker. It
// keeps the transactions table in sync with real on-chain outcomes for
// every transaction hash the indexer has observed, by polling Soroban RPC
// directly — see internal/reconcile.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	rpcclient "github.com/stellar/go-stellar-sdk/clients/rpcclient"

	nexusdb "github.com/StellarAsset-Lab/nexus-app/internal/db"
	"github.com/StellarAsset-Lab/nexus-app/internal/network"
	"github.com/StellarAsset-Lab/nexus-app/internal/reconcile"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "worker")

	if err := run(logger); err != nil {
		logger.Error("worker exited with error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is not set")
	}

	netConfig, err := network.LoadConfig()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conn, err := nexusdb.Open(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	rpc := rpcclient.NewClient(netConfig.RPCURL, http.DefaultClient)
	reconciler := reconcile.NewReconciler(rpc, conn, logger)

	interval := envDuration("WORKER_RECONCILE_INTERVAL", 10*time.Second)
	logger.Info("worker starting", "network", netConfig.Name, "reconcileInterval", interval)

	reconciler.RunLoop(ctx, interval)
	return nil
}

func envDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(v)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return time.Duration(parsed) * time.Second
}
