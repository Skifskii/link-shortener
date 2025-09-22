package app

import (
	"github.com/Skifskii/link-shortener/internal/config"
	"github.com/Skifskii/link-shortener/internal/logger"
	"github.com/Skifskii/link-shortener/internal/repository/file"
	"github.com/Skifskii/link-shortener/internal/router"
	"github.com/Skifskii/link-shortener/internal/service/shortener"
)

func Run() error {
	// Конфиг
	cfg := config.New()

	// Репозиторий
	repo, err := file.NewFileRepo(cfg.FileStoragePath)
	if err != nil {
		return err
	}

	// Логгер
	zl, err := logger.Init(cfg.LogLevel)
	if err != nil {
		return err
	}

	// Сервис сокращения ссылок
	s := shortener.New(cfg.BaseURL, 6, repo)

	// HTTP сервер
	r := router.New(zl, s)
	return r.Run(cfg.Address)
}
