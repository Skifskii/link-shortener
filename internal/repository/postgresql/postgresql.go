package postgresql

import (
	"database/sql"
	"errors"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var errEmptyDSN = errors.New("DSN is empty")

type PostgresqlRepo struct {
	db *sql.DB
}

func NewPostgresqlRepo(dsn string, zl *zap.Logger) (*PostgresqlRepo, error) {
	if dsn == "" {
		return nil, errEmptyDSN
	}

	// Запускаем миграции
	if err := runMigration(dsn, zl); err != nil {
		return nil, err
	}

	// Подключаемся к БД
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	return &PostgresqlRepo{db: db}, nil
}

func runMigration(dsn string, zl *zap.Logger) error {
	m, err := migrate.New(
		"file://./migrations",
		dsn,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	zl.Info("Migrations applied successfully!")
	return nil
}

func (pr *PostgresqlRepo) Save(short, original string) error {
	_, err := pr.db.Exec(
		"INSERT INTO links (short, original) VALUES ($1, $2)",
		short,
		original,
	)
	return err
}

func (pr *PostgresqlRepo) Get(short string) (original string, err error) {
	row := pr.db.QueryRow(
		"SELECT original FROM links WHERE short = $1 LIMIT 1",
		short,
	)

	err = row.Scan(&original)
	if err != nil {
		return "", err
	}

	return original, nil
}

func (pr *PostgresqlRepo) Ping() error {
	return pr.db.Ping()
}

func (pr *PostgresqlRepo) Close() error {
	return pr.db.Close()
}
