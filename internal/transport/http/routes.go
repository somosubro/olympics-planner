package http

import (
	"log/slog"
	nethttp "net/http"

	"olympics-planner/internal/config"
	"olympics-planner/internal/repository"
)

// NewRouter builds the /api/v1 mux with CORS for local browser testing.
func NewRouter(logger *slog.Logger, cfg config.Config) nethttp.Handler {
	sessRepo := repository.JSONSessionRepository{Path: cfg.SessionsFile}
	handler := NewHandler(logger, sessRepo)

	mux := nethttp.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", handler.Health)
	mux.HandleFunc("GET /api/v1/sessions", handler.ListSessions)
	mux.HandleFunc("POST /api/v1/validate", handler.Validate)
	mux.HandleFunc("POST /api/v1/rank/sessions", handler.RankSessions)
	mux.HandleFunc("POST /api/v1/rank/plans", handler.RankPlans)

	return withCORS(mux)
}
