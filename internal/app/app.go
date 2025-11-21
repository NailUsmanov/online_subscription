package app

import (
	"context"
	"net/http"
	"time"

	_ "github.com/NailUsmanov/online_subscription/internal/docs"
	"github.com/NailUsmanov/online_subscription/internal/handlers"
	"github.com/NailUsmanov/online_subscription/internal/middleware"
	"github.com/NailUsmanov/online_subscription/internal/service"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

// App - состоит из маршуртизатора chi, сервиса, логгера.
type App struct {
	router  *chi.Mux
	service service.Service
	sugar   *zap.SugaredLogger
}

// NewApp - создадим новую стркутуру Арр.
// В ней регистрируем маршруты.
func NewApp(s service.Service, sugar *zap.SugaredLogger) *App {
	r := chi.NewRouter()
	app := &App{
		router:  r,
		service: s,
		sugar:   sugar,
	}
	app.setupRoutes()
	return app
}

func (a *App) setupRoutes() {
	a.router.Use(middleware.LoggingMiddleware(a.sugar))

	a.router.Post("/subscriptions", handlers.CreateSubscriptionHandler(a.service, a.sugar))
	a.router.Get("/subscriptions/{id}", handlers.GetSubscriptionHandler(a.service, a.sugar))
	a.router.Put("/subscriptions/{id}", handlers.UpdateSubscriptionHandler(a.service, a.sugar))
	a.router.Delete("/subscriptions/{id}", handlers.DeleteSubscriptionsHandler(a.service, a.sugar))
	a.router.Get("/subscriptions", handlers.ListSubscriptionHandler(a.service, a.sugar))
	a.router.Get("/subscriptions/sum", handlers.SumSubscriptionHandler(a.service, a.sugar))

	a.router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))
}

// Run будет запускать HTTP-сервер на указаноом адресе
func (a *App) Run(ctx context.Context, addr string) error {
	srv := http.Server{
		Addr:    addr,
		Handler: a.router,
	}

	go func() {
		<-ctx.Done()
		a.sugar.Infof("Shutdown the server")
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
