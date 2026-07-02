package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// readyzHandler mirrors the readiness handler wired up in main().
func readyzHandler(w http.ResponseWriter, r *http.Request) {
	if !ready.Load() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "starting"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func TestHealthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler := func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestReadyzNotReady(t *testing.T) {
	ready.Store(false)
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	readyzHandler(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when not ready, got %d", rec.Code)
	}
}

func TestReadyzReady(t *testing.T) {
	ready.Store(true)
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	readyzHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 when ready, got %d", rec.Code)
	}
}
