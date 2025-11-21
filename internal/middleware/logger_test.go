package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestLoggingMiddleware_StatusAndSizeCaptured(t *testing.T) {
	logger := zap.NewNop().Sugar()

	var called bool

	// next-хендлер, который мы оборачиваем middleware
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("test-data"))
	})

	mw := LoggingMiddleware(logger)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	mw(next).ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if !called {
		t.Fatalf("next handler was not called")
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}

	body := w.Body.String()
	if body != "test-data" {
		t.Fatalf("body = %q, want %q", body, "test-data")
	}
}

func TestLoggingMiddleware_DefaultStatusOK(t *testing.T) {
	logger := zap.NewNop().Sugar()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ничего не пишем и не вызываем WriteHeader вручную.
		// Значит статус должен остаться 200.
	})

	mw := LoggingMiddleware(logger)

	req := httptest.NewRequest(http.MethodGet, "/empty", nil)
	w := httptest.NewRecorder()

	mw(next).ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}
