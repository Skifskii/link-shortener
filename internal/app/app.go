package app

import (
	"github.com/Skifskii/link-shortener/internal/config"
	"github.com/Skifskii/link-shortener/internal/logger"
	"github.com/Skifskii/link-shortener/internal/repository/file"
	"github.com/Skifskii/link-shortener/internal/repository/inmemory"
	"github.com/Skifskii/link-shortener/internal/repository/postgresql"
	"github.com/Skifskii/link-shortener/internal/router"
	"github.com/Skifskii/link-shortener/internal/service/dbping"
	"github.com/Skifskii/link-shortener/internal/service/shortener"
	"go.uber.org/zap"
)

func Run() error {
	// Конфиг
	cfg := config.New()

	// Логгер
	zl, err := logger.Init(cfg.LogLevel)
	if err != nil {
		return err
	}

	// Репозиторий
	var repo URLSaveGetter

	pgrepo, err := postgresql.NewPostgresqlRepo(cfg.DatabaseDSN)
	if err == nil {
		// Пробуем использовать Postgres
		defer pgrepo.Close()
		repo = pgrepo
		zl.Info("using postgresql as a storage")
	} else {
		// Ставим запасное хранилище
		zl.Warn("can't use postgresql as a storage: ", zap.Error(err))

		repo, err = chooseFallbackRepo(cfg, zl)
		if err != nil {
			return err
		}
	}

	// Сервис сокращения ссылок
	s := shortener.New(cfg.BaseURL, 6, repo)

	// Сервис проверки подключения к БД
	dBPingService := dbping.New(pgrepo)

	// HTTP сервер
	r := router.New(zl, s, dBPingService)
	return r.Run(cfg.Address)
}

type URLSaveGetter interface {
	Save(shortURL, longURL string) error
	Get(shortURL string) (string, error)
}

func chooseFallbackRepo(cfg *config.Config, zl *zap.Logger) (URLSaveGetter, error) {
	var repo URLSaveGetter
	var err error

	// Пробуем использовать файловую систему
	if repo, err = file.NewFileRepo(cfg.FileStoragePath); err == nil {
		zl.Info("using filesystem as a storage")
		return repo, nil
	}
	zl.Warn("can't use filesystem as a storage: ", zap.Error(err))

	// Используем хранение в памяти
	return inmemory.New(), nil
}
