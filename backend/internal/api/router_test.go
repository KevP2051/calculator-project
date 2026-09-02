package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testOrigin = "http://localhost:5173"

func do(t *testing.T, method, target string) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()
	NewRouter(testOrigin).ServeHTTP(recorder, httptest.NewRequest(method, target, nil))

	return recorder
}

func TestHealthReportsOK(t *testing.T) {
	recorder := do(t, http.MethodGet, "/healthz")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
	}

	var body map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status field = %q, want %q", body["status"], "ok")
	}
}

func TestResponsesCarryTheConfiguredCORSOrigin(t *testing.T) {
	recorder := httptest.NewRecorder()
	NewRouter("https://calculator.example").ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "https://calculator.example" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "https://calculator.example")
	}
}

func TestPreflightIsAnsweredWithoutContent(t *testing.T) {
	recorder := do(t, http.MethodOptions, "/api/v1/calculate")

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != testOrigin {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, testOrigin)
	}

	methods := recorder.Header().Get("Access-Control-Allow-Methods")
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodOptions} {
		if !strings.Contains(methods, method) {
			t.Errorf("Access-Control-Allow-Methods = %q, missing %s", methods, method)
		}
	}

	if got := recorder.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(got, "Content-Type") {
		t.Errorf("Access-Control-Allow-Headers = %q, missing Content-Type", got)
	}
}

func TestUnknownPathIsNotFound(t *testing.T) {
	if recorder := do(t, http.MethodGet, "/nope"); recorder.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestWrongMethodIsMethodNotAllowed(t *testing.T) {
	if recorder := do(t, http.MethodPost, "/healthz"); recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
}
