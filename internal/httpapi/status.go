package httpapi

import (
	"context"
	"net/http"
	"time"

	nexusdb "github.com/StellarAsset-Lab/nexus-app/internal/db"
	"github.com/StellarAsset-Lab/nexus-app/internal/model"
	"github.com/StellarAsset-Lab/nexus-app/internal/network"
)

type componentStatus struct {
	Name   string                `json:"name"`
	Status model.ComponentHealth `json:"status"`
	Detail string                `json:"detail,omitempty"`
}

type statusResponse struct {
	Components []componentStatus `json:"components"`
}

// handleStatus reports the real, currently-observed health of each
// dependency. A component is only ever "Operational" when its check
// actually succeeded this request — never defaulted, per model.ComponentHealth's
// contract.
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	components := []componentStatus{
		s.databaseStatus(ctx),
		s.rpcStatus(ctx),
		s.indexerStatus(ctx),
	}

	writeJSON(w, http.StatusOK, statusResponse{Components: components})
}

func (s *Server) databaseStatus(ctx context.Context) componentStatus {
	if err := s.db.PingContext(ctx); err != nil {
		return componentStatus{Name: "Database", Status: model.ComponentUnavailable, Detail: "The database is not reachable."}
	}
	return componentStatus{Name: "Database", Status: model.ComponentOperational}
}

func (s *Server) rpcStatus(ctx context.Context) componentStatus {
	if _, err := network.CheckRPCHealth(ctx, s.network.RPCURL); err != nil {
		return componentStatus{Name: "Soroban RPC", Status: model.ComponentUnavailable, Detail: "The configured RPC endpoint is not reachable."}
	}
	return componentStatus{Name: "Soroban RPC", Status: model.ComponentOperational}
}

// indexerLagThreshold is how long a checkpoint can go without advancing
// before the indexer is reported as Degraded rather than Operational.
const indexerLagThreshold = 5 * time.Minute

func (s *Server) indexerStatus(ctx context.Context) componentStatus {
	checkpoint, err := nexusdb.GetCheckpoint(ctx, s.db, string(s.network.Name))
	if err != nil {
		return componentStatus{Name: "Indexer", Status: model.ComponentUnknown, Detail: "Could not load the indexer checkpoint."}
	}
	if checkpoint == nil {
		return componentStatus{Name: "Indexer", Status: model.ComponentUnknown, Detail: "The indexer has not recorded a checkpoint yet."}
	}
	if time.Since(checkpoint.UpdatedAt) > indexerLagThreshold {
		return componentStatus{Name: "Indexer", Status: model.ComponentDegraded, Detail: "The indexer checkpoint has not advanced recently."}
	}
	return componentStatus{Name: "Indexer", Status: model.ComponentOperational}
}
