package app

import (
	"github.com/Skifskii/link-shortener/internal/config"
	"github.com/Skifskii/link-shortener/internal/logger"
	"github.com/Skifskii/link-shortener/internal/repository/file"
	"github.com/Skifskii/link-shortener/internal/repository/postgresql"
	"github.com/Skifskii/link-shortener/internal/router"
	"github.com/Skifskii/link-shortener/internal/service/dbping"
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

	pgrepo, err := postgresql.NewPostgresqlRepo(cfg.DatabaseDSN)
	if err != nil {
		return err
	}
	defer pgrepo.Close()

	// Логгер
	zl, err := logger.Init(cfg.LogLevel)
	if err != nil {
		return err
	}

	// Сервис сокращения ссылок
	s := shortener.New(cfg.BaseURL, 6, repo)

	// Сервис проверки подключения к БД
	dBPingService := dbping.New(pgrepo)

	// HTTP сервер
	r := router.New(zl, s, dBPingService)
	return r.Run(cfg.Address)
}
