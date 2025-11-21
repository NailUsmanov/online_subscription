package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NailUsmanov/online_subscription/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func newTestLogger(t *testing.T) *zap.SugaredLogger {
	t.Helper()
	l := zap.NewNop()
	return l.Sugar()
}

func withRouteParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	return req.WithContext(ctx)
}

func TestCreateSubscriptionHandler_BadJSON(t *testing.T) {
	var svc service.Service
	logger := newTestLogger(t)

	h := CreateSubscriptionHandler(svc, logger)

	req := httptest.NewRequest(http.MethodPost, "/subscriptions", bytes.NewBufferString("{bad"))
	w := httptest.NewRecorder()

	h(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	got := w.Body.String()
	want := "Invalid JSON format\n"
	if got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestGetSubscriptionHandler_NoID(t *testing.T) {
	var svc service.Service
	logger := newTestLogger(t)

	h := GetSubscriptionHandler(svc, logger)

	req := httptest.NewRequest(http.MethodGet, "/subscriptions", nil)
	w := httptest.NewRecorder()

	h(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	got := w.Body.String()
	want := "invalid request id\n"
	if got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestGetSubscriptionHandler_InvalidID(t *testing.T) {
	var svc service.Service
	logger := newTestLogger(t)

	h := GetSubscriptionHandler(svc, logger)

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/abc", nil)
	req = withRouteParam(req, "id", "abc")

	w := httptest.NewRecorder()
	h(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	got := w.Body.String()
	want := "id must be integer\n"
	if got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestUpdateSubscriptionHandler_NoID(t *testing.T) {
	var svc service.Service
	logger := newTestLogger(t)

	h := UpdateSubscriptionHandler(svc, logger)

	req := httptest.NewRequest(http.MethodPut, "/subscriptions", bytes.NewBufferString(`{}`))
	w := httptest.NewRecorder()

	h(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	got := w.Body.String()
	want := "invalid id\n"
	if got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestUpdateSubscriptionHandler_InvalidID(t *testing.T) {
	var svc service.Service
	logger := newTestLogger(t)

	h := UpdateSubscriptionHandler(svc, logger)

	req := httptest.NewRequest(http.MethodPut, "/subscriptions/abc", bytes.NewBufferString(`{}`))
	req = withRouteParam(req, "id", "abc")

	w := httptest.NewRecorder()
	h(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	got := w.Body.String()
	want := "id must be integer\n"
	if got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestUpdateSubscriptionHandler_BadJSON(t *testing.T) {
	var svc service.Service
	logger := newTestLogger(t)

	h := UpdateSubscriptionHandler(svc, logger)

	req := httptest.NewRequest(http.MethodPut, "/subscriptions/1", bytes.NewBufferString("{bad"))
	req = withRouteParam(req, "id", "1")

	w := httptest.NewRecorder()
	h(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	got := w.Body.String()
	want := "Invalid JSON format\n"
	if got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestDeleteSubscriptionsHandler_NoID(t *testing.T) {
	var svc service.Service
	logger := newTestLogger(t)

	h := DeleteSubscriptionsHandler(svc, logger)

	req := httptest.NewRequest(http.MethodDelete, "/subscriptions", nil)
	w := httptest.NewRecorder()

	h(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	got := w.Body.String()
	want := "invalid request id\n"
	if got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestDeleteSubscriptionsHandler_InvalidID(t *testing.T) {
	var svc service.Service
	logger := newTestLogger(t)

	h := DeleteSubscriptionsHandler(svc, logger)

	req := httptest.NewRequest(http.MethodDelete, "/subscriptions/abc", nil)
	req = withRouteParam(req, "id", "abc")

	w := httptest.NewRecorder()
	h(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	got := w.Body.String()
	want := "id must be integer\n"
	if got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestListSubscriptionHandler_NoUserID(t *testing.T) {
	var svc service.Service
	logger := newTestLogger(t)

	h := ListSubscriptionHandler(svc, logger)

	req := httptest.NewRequest(http.MethodGet, "/subscriptions", nil)
	w := httptest.NewRecorder()

	h(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	got := w.Body.String()
	want := "invalid userID in request\n"
	if got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestSumSubscriptionHandler_MissingParams(t *testing.T) {
	var svc service.Service
	logger := newTestLogger(t)

	h := SumSubscriptionHandler(svc, logger)

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/sum", nil)
	w := httptest.NewRecorder()

	h(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	got := w.Body.String()
	want := "user_id, service_name, from, to are required\n"
	if got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}
