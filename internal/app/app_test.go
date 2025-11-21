package app

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/NailUsmanov/online_subscription/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// newTestApp - вспомогательная функция для тестов.
func newTestApp(t *testing.T) *App {
	t.Helper()

	var s service.Service

	logger := zap.NewNop() // логгер без вывода
	sugar := logger.Sugar()
	return NewApp(s, sugar)
}

// TestAppRoutesRegistered проверяет, что все ожидаемые маршруты
// реально зарегистрированы в chi.Router через setupRoutes.
func TestAppRoutesRegistered(t *testing.T) {
	app := newTestApp(t)

	expected := map[string]map[string]bool{
		http.MethodPost: {
			"/subscriptions": true,
		},
		http.MethodGet: {
			"/subscriptions/{id}": true,
			"/subscriptions":      true,
			"/subscriptions/sum":  true,
			"/swagger/*":          true,
		},
		http.MethodPut: {
			"/subscriptions/{id}": true,
		},
		http.MethodDelete: {
			"/subscriptions/{id}": true,
		},
	}

	// обходим все маршруты chi с помощью Walk
	err := chi.Walk(app.router, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if m, ok := expected[method]; ok {
			if _, ok := m[route]; ok {
				m[route] = false
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("chi.Walk failed: %v", err)
	}

	for method, routes := range expected {
		for route, stillExpected := range routes {
			if stillExpected {
				t.Errorf("route not registered: %s %s", method, route)
			}
		}
	}
}

func TestAppRun_ShutdownOnContextCancel(t *testing.T) {
	app := newTestApp(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Run(ctx, "127.0.0.1:0")
	}()

	time.Sleep(100 * time.Millisecond)

	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Run returned error after context cancel: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after context cancellation")
	}
}
