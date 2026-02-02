package main

import (
	"api-project/internal/auth"
	"api-project/internal/config"
	myhttp "api-project/internal/http"
	"api-project/internal/storage"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
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

	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) < 16 {
		logger.Error("JWT_SECRET is too short (set env JWT_SECRET)")
		return
	}

	handlers := myhttp.New(store, logger, jwtSecret)

	// ВРЕМЕННО: вывести токен в лог, чтобы руками проверить middleware
	t, _ := auth.GenerateAccessToken(1, jwtSecret, 15*time.Minute)
	logger.Info("DEV TOKEN", slog.String("token", t))

	router := chi.NewRouter()

	router.Use(middleware.RequestID)                 // добавляет request-id в контекст
	router.Use(middleware.Logger)                    // пишет в лог метод/путь/время
	router.Use(middleware.Recoverer)                 // ловит panic и возвращает 500
	router.Use(middleware.Timeout(10 * time.Second)) // чтобы запрос не висел вечно

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	router.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", handlers.Login)
			r.Post("/refresh", handlers.Refresh)
			r.Post("/logout", handlers.Logout)
		})

		r.Group(func(r chi.Router) {
			r.Use(handlers.AuthRequired)
			r.Post("/url", handlers.CreateURL)
			r.Delete("/url/{alias}", handlers.DeleteURL)
		})
	})

	router.Get("/{alias}", handlers.Redirect)

	logger.Info("starting http server", slog.String("addr", ":8080"))
	if err := http.ListenAndServe(":8080", router); err != nil {
		logger.Error("server stopped", slog.Any("err", err))
	}
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
