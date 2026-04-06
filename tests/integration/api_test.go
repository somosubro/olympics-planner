package integration

import (
    "log/slog"
    "net/http"
    "net/http/httptest"
    "testing"

    apphttp "olympics-planner/internal/transport/http"
)

func TestHealthEndpoint(t *testing.T) {
    router := apphttp.NewRouter(slog.Default())

    req := httptest.NewRequest(http.MethodGet, "/health", nil)
    rr := httptest.NewRecorder()

    router.ServeHTTP(rr, req)

    if rr.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rr.Code)
    }
}
