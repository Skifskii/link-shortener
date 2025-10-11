package postgresql

import (
	"database/sql"
	"errors"

	"github.com/Skifskii/link-shortener/internal/model"
	"github.com/Skifskii/link-shortener/internal/repository"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var errEmptyDSN = errors.New("DSN is empty")
var errDifferentSliceSizes = errors.New("slices are of different sizes")
var errEmptyBatch = errors.New("batch is empty")

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

func (pr *PostgresqlRepo) Save(userID int, short, original string) (savedShort string, err error) {
	var linkID int

	err = pr.db.QueryRow(
		`INSERT INTO links (short, original)
		VALUES ($1, $2)
		ON CONFLICT (original) DO UPDATE
			SET short = links.short
		RETURNING id, short`,
		short, original,
	).Scan(&linkID, &savedShort)
	if err != nil {
		return "", err
	}

	_, err = pr.db.Exec(
		`INSERT INTO users_links (user_id, link_id)
		VALUES ($1, $2)`,
		userID, linkID,
	)
	if err != nil {
		return "", err
	}

	if savedShort != short {
		return savedShort, repository.ErrOriginalURLAlreadyExists
	}

	return "", nil
}

func (pr *PostgresqlRepo) SaveBatch(shortURLs, longURLs []string) error {
	if len(shortURLs) != len(longURLs) {
		return errDifferentSliceSizes
	}

	if len(shortURLs) == 0 {
		return errEmptyBatch
	}

	tx, err := pr.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO links (short, original) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i, short := range shortURLs {
		_, err := stmt.Exec(short, longURLs[i])
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (pr *PostgresqlRepo) Get(short string) (original string, err error) {
	row := pr.db.QueryRow(
		"SELECT original, is_deleted FROM links WHERE short = $1 LIMIT 1",
		short,
	)

	var isDeleted bool
	err = row.Scan(&original, &isDeleted)
	if err != nil {
		return "", err
	}

	if isDeleted {
		return "", repository.ErrLinkDeleted
	}

	return original, nil
}

func (pr *PostgresqlRepo) Ping() error {
	return pr.db.Ping()
}

func (pr *PostgresqlRepo) Close() error {
	return pr.db.Close()
}

func (pr *PostgresqlRepo) CreateUser(username string) (userID int, err error) {
	err = pr.db.QueryRow(
		`INSERT INTO users (username)
		VALUES ($1)
		RETURNING id`,
		username,
	).Scan(&userID)

	return userID, err
}

func (pr *PostgresqlRepo) GetUserPairs(userID int) ([]model.ResponsePairElement, error) {
	rows, err := pr.db.Query(
		`SELECT l.short, l.original
		FROM links AS l
		JOIN users_links AS ul ON ul.link_id = l.id
		WHERE ul.user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pairs := make([]model.ResponsePairElement, 0)
	for rows.Next() {
		var e model.ResponsePairElement
		if err := rows.Scan(&e.ShortURL, &e.OriginalURL); err != nil {
			return nil, err
		}

		pairs = append(pairs, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pairs, nil
}

func (pr *PostgresqlRepo) DeleteLinkByShort(userID int, shortURL string) error {
	_, err := pr.db.Exec(
		`UPDATE links
		SET is_deleted = TRUE
		WHERE id = (
			SELECT l.id
			FROM links AS l
			JOIN users_links AS ul ON ul.link_id = l.id
			WHERE l.short = $1 AND ul.user_id = $2
			LIMIT 1
		);`,
		shortURL, userID,
	)
	if err != nil {
		return err
	}

	return nil
}
