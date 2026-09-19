// Package indexing implements the Nexus indexer's RPC polling, checkpoint
// resolution, and (in later commits) event decoding and projection logic.
package indexing

import (
	"context"
	"database/sql"
	"fmt"

	nexusdb "github.com/StellarAsset-Lab/nexus-app/internal/db"
)

// CheckpointID identifies the one logical checkpoint for a given indexed
// network. One checkpoint per network/configuration, per spec §15.
func CheckpointID(networkName string) string {
	return networkName
}

// ResolveStartLedger determines the safe ledger to resume indexing from.
//
// The persisted checkpoint always takes precedence once the indexer has
// begun successfully. configuredStartLedger (INDEXER_START_LEDGER) only
// seeds a first deployment, before any checkpoint exists. In every case the
// result is clamped to the ledger range Stellar RPC actually still retains —
// RPC event history is bounded, so a stale checkpoint older than the
// retention window cannot be resumed from literally; it resumes from the
// oldest ledger RPC still has.
func ResolveStartLedger(
	ctx context.Context,
	q *sql.DB,
	networkName string,
	configuredStartLedger *uint32,
	oldestLedger, latestLedger uint32,
) (uint32, error) {
	checkpoint, err := nexusdb.GetCheckpoint(ctx, q, CheckpointID(networkName))
	if err != nil {
		return 0, fmt.Errorf("load checkpoint: %w", err)
	}

	var start uint32
	switch {
	case checkpoint != nil:
		start = uint32(checkpoint.LastLedger) + 1
	case configuredStartLedger != nil:
		start = *configuredStartLedger
	default:
		start = oldestLedger
	}

	if start < oldestLedger {
		start = oldestLedger
	}
	if start > latestLedger {
		start = latestLedger
	}

	return start, nil
}
