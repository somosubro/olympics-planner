package http

import (
    "encoding/json"
    "log/slog"
    nethttp "net/http"
)

type Handler struct {
    logger *slog.Logger
}

func NewHandler(logger *slog.Logger) *Handler {
    return &Handler{logger: logger}
}

func (h *Handler) Health(w nethttp.ResponseWriter, _ *nethttp.Request) {
    writeJSON(w, nethttp.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w nethttp.ResponseWriter, status int, payload interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(payload)
}
