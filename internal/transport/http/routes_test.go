package http

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"olympics-planner/internal/config"
)

func testConfig() config.Config {
	return config.Config{SessionsFile: "../../../testdata/sessions_small.json"}
}

func TestRouter_HealthAPIv1(t *testing.T) {
	r := NewRouter(slog.Default(), testConfig())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("body %#v", body)
	}
}

func TestRouter_ForbiddenOrchestrationRoutesNotRegistered(t *testing.T) {
	r := NewRouter(slog.Default(), testConfig())
	cases := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/generate-weekend-plan"},
		{http.MethodPost, "/api/v1/generate-saturday-plan"},
		{http.MethodPost, "/api/v1/generate-multi-day-plan"},
		{http.MethodPost, "/api/v1/best-saturday-plan"},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			b, _ := io.ReadAll(rr.Body)
			t.Fatalf("%s %s: expected 404, got %d body=%s", tc.method, tc.path, rr.Code, b)
		}
	}
}
