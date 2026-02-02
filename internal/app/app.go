// Package app собирает и запускает все компоненты приложения: конфигурацию,
// логгер, репозиторий, сервисы и HTTP роутер.
package app

import (
	"context"
	"fmt"

	"github.com/Skifskii/link-shortener/internal/config"
	"github.com/Skifskii/link-shortener/internal/logger"
	"github.com/Skifskii/link-shortener/internal/model"
	"github.com/Skifskii/link-shortener/internal/repository/file"
	"github.com/Skifskii/link-shortener/internal/repository/inmemory"
	"github.com/Skifskii/link-shortener/internal/repository/postgresql"
	"github.com/Skifskii/link-shortener/internal/router"
	"github.com/Skifskii/link-shortener/internal/service/audit"
	"github.com/Skifskii/link-shortener/internal/service/audit/fileobs"
	"github.com/Skifskii/link-shortener/internal/service/audit/urlobs"
	"github.com/Skifskii/link-shortener/internal/service/auth"
	"github.com/Skifskii/link-shortener/internal/service/dbping"
	"github.com/Skifskii/link-shortener/internal/service/shortener"
	"github.com/Skifskii/link-shortener/internal/service/stats"
	grpcserver "github.com/Skifskii/link-shortener/internal/transport/grpc/server"
	"golang.org/x/sync/errgroup"

	"go.uber.org/zap"
)

// Run инициализирует все компоненты приложения и запускает HTTP-сервер.
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
	pgrepo, err := postgresql.NewPostgresqlRepo(cfg.DatabaseDSN, zl)
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

	// ===== Сервисы =====
	// - сервис сокращения ссылок
	s := shortener.New(cfg.BaseURL, 6, repo)
	// - сервис проверки подключения к БД
	dBPingService := dbping.New(pgrepo)
	// - сервис аутентификации
	authServiece := auth.New(repo, cfg.SecretKey)
	// - сервис аудита запросов
	auditService := audit.New()
	if cfg.AuditFile != "" {
		auditService.Register(fileobs.New(cfg.AuditFile))
	}
	if cfg.AuditURL != "" {
		auditService.Register(urlobs.New(cfg.AuditURL))
	}
	// - сервис статистики
	statsService, err := stats.New(cfg.TrustedSubnet, repo)
	if err != nil {
		return fmt.Errorf("failed to initialize stats service: %w", err)
	}

	// ===== Транспортный слой =====
	g, ctx := errgroup.WithContext(context.Background())
	// - HTTP сервер
	r := router.New(zl, s, dBPingService, authServiece, auditService, statsService)
	g.Go(func() error {
		return r.Run(zl, ctx, cfg.Address, cfg.TLSCertPath, cfg.TLSKeyPath, cfg.EnableHTTPS)
	})
	// - gRPC сервер
	grpcServer := grpcserver.New(s, authServiece, s, s)
	g.Go(func() error {
		return grpcServer.Run(ctx, cfg.GRPCPort)
	})

	return g.Wait()
}

// URLSaveGetter определяет набор методов, которые приложение ожидает от
// репозитория для сохранения и получения URL-ов (используется и в app, и в сервисе сокращения).
type URLSaveGetter interface {
	Save(userID int, shortURL, longURL string) (existingShort string, err error)
	Get(shortURL string) (string, error)
	SaveBatch(userID int, shortURLs, longURLs []string) error
	GetUserPairs(userID int) ([]model.ResponsePairElement, error)
	CreateUser(username string) (userID int, err error)
	DeleteBatchOfLinks(userID int, shortURL []string) error
	GetUsersCount() (int, error)
	GetURLsCount() (int, error)
}

// chooseFallbackRepo выбирает запасное хранилище (файл или память) и возвращает его.
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
