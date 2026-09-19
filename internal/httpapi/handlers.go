package httpapi

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	nexusdb "github.com/StellarAsset-Lab/nexus-app/internal/db"
	"github.com/StellarAsset-Lab/nexus-app/internal/network"
)

// Server holds the dependencies the API's handlers need. It never holds a
// wallet key or signs transactions — the API is read-oriented over indexed
// PostgreSQL data (spec §14).
type Server struct {
	logger  *slog.Logger
	db      *sql.DB
	network network.Config
}

func NewServer(logger *slog.Logger, db *sql.DB, netConfig network.Config) *Server {
	return &Server{logger: logger, db: db, network: netConfig}
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		s.logger.Error("readiness check failed: database unreachable", "error", err)
		writeError(w, http.StatusServiceUnavailable, "NOT_READY", "The database is not reachable.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

type networkResponse struct {
	Network           string  `json:"network"`
	NetworkPassphrase string  `json:"networkPassphrase"`
	RPCHealthy        bool    `json:"rpcHealthy"`
	LatestLedger      *uint32 `json:"latestLedger,omitempty"`
	OldestLedger      *uint32 `json:"oldestLedger,omitempty"`
	IndexerCheckpoint *int64  `json:"indexerCheckpoint,omitempty"`
}

func (s *Server) handleNetwork(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	resp := networkResponse{
		Network:           string(s.network.Name),
		NetworkPassphrase: s.network.Passphrase,
	}

	if health, err := network.CheckRPCHealth(ctx, s.network.RPCURL); err != nil {
		s.logger.Warn("rpc health check failed", "error", err)
		resp.RPCHealthy = false
	} else {
		resp.RPCHealthy = true
		resp.LatestLedger = &health.LatestLedger
		resp.OldestLedger = &health.OldestLedger
	}

	checkpoint, err := nexusdb.GetCheckpoint(ctx, s.db, string(s.network.Name))
	if err != nil {
		s.logger.Error("failed to load indexer checkpoint", "error", err)
	} else if checkpoint != nil {
		resp.IndexerCheckpoint = &checkpoint.LastLedger
	}

	writeJSON(w, http.StatusOK, resp)
}
