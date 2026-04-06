package http

import (
    "log/slog"
    nethttp "net/http"
)

func NewRouter(logger *slog.Logger) nethttp.Handler {
    handler := NewHandler(logger)

    mux := nethttp.NewServeMux()
    mux.HandleFunc("GET /health", handler.Health)

    return mux
}
