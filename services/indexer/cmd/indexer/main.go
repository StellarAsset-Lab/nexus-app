// Command indexer is the entry point for the Nexus Stellar event indexer.
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

	rpcclient "github.com/stellar/go-stellar-sdk/clients/rpcclient"

	nexusdb "github.com/StellarAsset-Lab/nexus-app/internal/db"
	"github.com/StellarAsset-Lab/nexus-app/internal/indexing"
	"github.com/StellarAsset-Lab/nexus-app/internal/network"
)

const (
	contractIDPlaceholder  = "<SET_AFTER_DEPLOYMENT>"
	startLedgerPlaceholder = "<OPTIONAL_START_LEDGER>"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "indexer")

	if err := run(logger); err != nil {
		logger.Error("indexer exited with error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	registryContractID := resolveContractID(os.Getenv("NEXT_PUBLIC_REGISTRY_CONTRACT_ID"))
	orderContractID := resolveContractID(os.Getenv("NEXT_PUBLIC_ORDER_CONTRACT_ID"))

	if registryContractID == "" && orderContractID == "" {
		logger.Info("no Registry or Order contract ID configured; nothing to index until Phase 8 deployment")
		return nil
	}

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

	poller, err := indexing.NewPoller(rpc, conn, logger, indexing.PollerConfig{
		NetworkName:           string(netConfig.Name),
		RegistryContractID:    registryContractID,
		OrderContractID:       orderContractID,
		BatchLimit:            envUint("INDEXER_BATCH_LIMIT", 100),
		ConfirmationLookback:  uint32(envUint("INDEXER_CONFIRMATION_LOOKBACK", 32)),
		ConfiguredStartLedger: resolveStartLedger(os.Getenv("INDEXER_START_LEDGER")),
	})
	if err != nil {
		return err
	}

	logger.Info("indexer starting",
		"network", netConfig.Name,
		"registryContract", registryContractID,
		"orderContract", orderContractID,
	)

	return poller.Run(ctx)
}

func resolveContractID(v string) string {
	if v == "" || v == contractIDPlaceholder {
		return ""
	}
	return v
}

func resolveStartLedger(v string) *uint32 {
	if v == "" || v == startLedgerPlaceholder {
		return nil
	}
	parsed, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return nil
	}
	result := uint32(parsed)
	return &result
}

func envUint(key string, fallback uint) uint {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return fallback
	}
	return uint(parsed)
}
