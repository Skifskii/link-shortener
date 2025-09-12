package app

import (
	"fmt"
	"net/http"

	"github.com/Skifskii/link-shortener/internal/config"
	"github.com/Skifskii/link-shortener/internal/handler/redirect"
	"github.com/Skifskii/link-shortener/internal/handler/save"
	"github.com/Skifskii/link-shortener/internal/handler/shorten"
	"github.com/Skifskii/link-shortener/internal/logger"
	"github.com/Skifskii/link-shortener/internal/middleware"
	"github.com/Skifskii/link-shortener/internal/repository"
	"github.com/Skifskii/link-shortener/internal/repository/inmemory"
	"github.com/Skifskii/link-shortener/internal/service/shortener"
	"github.com/go-chi/chi/v5"
)

func Run() error {
	// Конфиг
	cfg := config.New()

	// Репозиторий
	var repo repository.Repository = inmemory.New()

	// Сервис сокращения ссылок
	s := shortener.New(6)

	// Логгер
	zl, err := logger.Init(cfg.LogLevel)
	if err != nil {
		return err
	}

	// HTTP сервер
	r := chi.NewRouter()
	r.Use(logger.RequestLogger(zl))
	r.Use(middleware.GzipMiddleware)
	r.Get("/{id}", redirect.New(repo))
	r.Post("/", save.New(repo, s, cfg.BaseURL))
	r.Post("/api/shorten", shorten.New(repo, s, cfg.BaseURL))

	fmt.Printf("Starting server at %s\n", cfg.Address)
	return http.ListenAndServe(cfg.Address, r)
}
