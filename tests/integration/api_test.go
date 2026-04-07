package integration

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"olympics-planner/internal/config"
	apphttp "olympics-planner/internal/transport/http"
)

func TestHealthEndpoint(t *testing.T) {
	cfg := config.Config{SessionsFile: "../../testdata/sessions_small.json"}
	router := apphttp.NewRouter(slog.Default(), cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
