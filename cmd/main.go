package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/NailUsmanov/online_subscription/internal/app"
	"github.com/NailUsmanov/online_subscription/internal/config"
	"github.com/NailUsmanov/online_subscription/internal/service"
	"github.com/NailUsmanov/online_subscription/internal/storage"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// проглатываем ошибку, чтобы не падать, если файла нет
	_ = godotenv.Load()
	// Запуск логгера
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	// задаю обертку над логгером
	sugar := logger.Sugar()

	// чтение конфига
	cfg, err := config.Load()
	if err != nil {
		sugar.Fatalw("failed to load config", "error", err)
	}

	// контекст для Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// хранилище для ссылок
	storage, err := storage.NewStorage(cfg.DatabaseDSN)
	if err != nil {
		sugar.Fatalf("failed to initialization DataBase: %v", err)
	}

	// Создаем сервис.
	svc := service.NewService(storage)

	// создаем арр
	application := app.NewApp(svc, sugar)

	// Логирую запуск сервера и вызывваю Run
	sugar.Infow("Starting HTTP server", "addr", cfg.ServerAddr)
	if err := application.Run(ctx, cfg.ServerAddr); err != nil {
		sugar.Fatalln(err)
	}
	sugar.Infow("server stop")
}
