package httpapi

import "net/http"

// Routes builds the API's HTTP handler. Uses net/http's Go 1.22+
// method-and-path pattern matching directly — no router dependency needed
// for this route set.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /readyz", s.handleReadyz)
	mux.HandleFunc("GET /api/v1/network", s.handleNetwork)
	mux.HandleFunc("GET /api/v1/assets", s.handleListAssets)
	mux.HandleFunc("GET /api/v1/assets/{asset}", s.handleGetAsset)
	mux.HandleFunc("GET /api/v1/assets/{asset}/activity", s.handleAssetActivity)
	mux.HandleFunc("GET /api/v1/orders", s.handleListOrders)
	mux.HandleFunc("GET /api/v1/orders/{id}", s.handleGetOrder)
	mux.HandleFunc("GET /api/v1/transactions/{hash}", s.handleGetTransaction)
	mux.HandleFunc("GET /api/v1/events", s.handleListEvents)
	mux.HandleFunc("GET /api/v1/status", s.handleStatus)

	return withRecovery(s.logger, withLogging(s.logger, mux))
}
