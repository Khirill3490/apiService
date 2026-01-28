package main

import (
	"api-project/internal/config"
	"api-project/internal/storage"
	"fmt"
	"log/slog"
	"os"
)

func main() {
	cfg := config.MustLoad()
	logger := setupLogger(cfg.Env)

	logger.Info("Приложение запущено", slog.String("env", cfg.Env))

	store, err := storage.NewPostgres()
	if err != nil {
		logger.Error("Ошибка подключения к базе данных", slog.String("error", err.Error()))
		return
	}
	defer store.Close()

	logger.Info("Подключение к базе данных успешно установлено")

	fmt.Printf("Конфигурация загружена: %+v\n", cfg)
}

const (
	envLocal = "local"
	envDev   = "development"
	envProd  = "production"
)

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	default:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
