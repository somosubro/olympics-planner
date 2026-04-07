package http

import (
	"encoding/json"
	"io"
	"log/slog"
	nethttp "net/http"

	"olympics-planner/internal/domain"
	"olympics-planner/internal/planner"
	"olympics-planner/internal/repository"
	sess "olympics-planner/internal/session"
)

// Handler serves the MVP API (api-spec.md).
type Handler struct {
	logger   *slog.Logger
	sessions repository.SessionRepository
}

// NewHandler constructs a Handler with required dependencies.
func NewHandler(logger *slog.Logger, sessions repository.SessionRepository) *Handler {
	return &Handler{logger: logger, sessions: sessions}
}

// Health handles GET /api/v1/health.
func (h *Handler) Health(w nethttp.ResponseWriter, _ *nethttp.Request) {
	writeJSON(w, nethttp.StatusOK, map[string]string{"status": "ok"})
}

// ListSessions handles GET /api/v1/sessions.
func (h *Handler) ListSessions(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodGet {
		httpError(w, nethttp.StatusMethodNotAllowed, "method not allowed")
		return
	}
	f, errBody := ParseSessionsFilter(r)
	if errBody != nil {
		writeError(w, nethttp.StatusBadRequest, errBody)
		return
	}
	all, err := h.sessions.GetAll()
	if err != nil {
		h.logger.Error("load sessions", "error", err)
		writeError(w, nethttp.StatusInternalServerError, &ErrorBody{Code: "INTERNAL_ERROR", Message: "failed to load session dataset"})
		return
	}
	out := sess.ApplyFilter(all, f)
	writeJSON(w, nethttp.StatusOK, map[string]interface{}{"sessions": out})
}

type validateRequest struct {
	Plan        domain.Plan        `json:"plan"`
	Preferences domain.Preferences `json:"preferences"`
}

// Validate handles POST /api/v1/validate.
func (h *Handler) Validate(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodPost {
		httpError(w, nethttp.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req validateRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, nethttp.StatusBadRequest, &ErrorBody{Code: "INVALID_REQUEST_BODY", Message: err.Error(), Field: "body"})
		return
	}
	if req.Plan.PlanType == "" || len(req.Plan.Days) == 0 {
		writeError(w, nethttp.StatusBadRequest, &ErrorBody{Code: "INVALID_REQUEST_BODY", Message: "plan is required", Field: "plan"})
		return
	}
	dataset, err := h.sessions.GetAll()
	if err != nil {
		h.logger.Error("load sessions", "error", err)
		writeError(w, nethttp.StatusInternalServerError, &ErrorBody{Code: "INTERNAL_ERROR", Message: "failed to load session dataset"})
		return
	}
	res := planner.ValidatePlan(req.Plan, dataset, req.Preferences)
	writeJSON(w, nethttp.StatusOK, res)
}

func prefsLooksPresent(p domain.Preferences) bool {
	return len(p.SportPriority) > 0 || len(p.AllowedSports) > 0 || len(p.AllowedDays) > 0
}

type rankSessionsRequest struct {
	Sessions              []domain.Session   `json:"sessions"`
	Preferences           domain.Preferences `json:"preferences"`
	IncludeScoreBreakdown *bool              `json:"includeScoreBreakdown"`
}

// RankSessions handles POST /api/v1/rank/sessions.
func (h *Handler) RankSessions(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodPost {
		httpError(w, nethttp.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req rankSessionsRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, nethttp.StatusBadRequest, &ErrorBody{Code: "INVALID_REQUEST_BODY", Message: err.Error(), Field: "body"})
		return
	}
	if len(req.Sessions) == 0 {
		writeError(w, nethttp.StatusBadRequest, &ErrorBody{Code: "INVALID_REQUEST_BODY", Message: "sessions is required", Field: "sessions"})
		return
	}
	if !prefsLooksPresent(req.Preferences) {
		writeError(w, nethttp.StatusBadRequest, &ErrorBody{Code: "INVALID_REQUEST_BODY", Message: "preferences is required", Field: "preferences"})
		return
	}
	breakdown := req.IncludeScoreBreakdown != nil && *req.IncludeScoreBreakdown
	rows := planner.RankSessions(req.Sessions, req.Preferences, breakdown)
	writeJSON(w, nethttp.StatusOK, map[string]interface{}{"rankedSessions": rows})
}

type rankPlansRequest struct {
	Plans                 []domain.Plan      `json:"plans"`
	Preferences           domain.Preferences `json:"preferences"`
	IncludeScoreBreakdown *bool              `json:"includeScoreBreakdown"`
	IncludeInvalidPlans   *bool              `json:"includeInvalidPlans"`
}

// RankPlans handles POST /api/v1/rank/plans.
func (h *Handler) RankPlans(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodPost {
		httpError(w, nethttp.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req rankPlansRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, nethttp.StatusBadRequest, &ErrorBody{Code: "INVALID_REQUEST_BODY", Message: err.Error(), Field: "body"})
		return
	}
	if len(req.Plans) == 0 {
		writeError(w, nethttp.StatusBadRequest, &ErrorBody{Code: "INVALID_REQUEST_BODY", Message: "plans is required", Field: "plans"})
		return
	}
	if !prefsLooksPresent(req.Preferences) {
		writeError(w, nethttp.StatusBadRequest, &ErrorBody{Code: "INVALID_REQUEST_BODY", Message: "preferences is required", Field: "preferences"})
		return
	}
	dataset, err := h.sessions.GetAll()
	if err != nil {
		h.logger.Error("load sessions", "error", err)
		writeError(w, nethttp.StatusInternalServerError, &ErrorBody{Code: "INTERNAL_ERROR", Message: "failed to load session dataset"})
		return
	}
	breakdown := req.IncludeScoreBreakdown != nil && *req.IncludeScoreBreakdown
	inv := req.IncludeInvalidPlans != nil && *req.IncludeInvalidPlans
	resp := planner.RankPlans(req.Plans, dataset, req.Preferences, breakdown, inv)
	writeJSON(w, nethttp.StatusOK, resp)
}

func decodeJSON(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

func writeJSON(w nethttp.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w nethttp.ResponseWriter, status int, body *ErrorBody) {
	writeJSON(w, status, errorEnvelope{Error: *body})
}

func httpError(w nethttp.ResponseWriter, status int, msg string) {
	nethttp.Error(w, msg, status)
}
