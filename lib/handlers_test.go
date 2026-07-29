package lib

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestReadyzChecksDependencies(t *testing.T) {
	ready := &atomic.Value{}
	ready.Store(true)

	handler := Readyz(ready, func(context.Context) error {
		return errors.New("database unavailable")
	})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/ready", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", recorder.Code)
	}
}

func TestReadyzReturnsOKWhenReady(t *testing.T) {
	ready := &atomic.Value{}
	ready.Store(true)

	handler := Readyz(ready, func(context.Context) error { return nil })
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/ready", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}
