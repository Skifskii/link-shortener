package postgresql

import (
	"database/sql"
	"errors"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var errEmptyDSN = errors.New("DSN is empty")

type PostgresqlRepo struct {
	db *sql.DB
}

func NewPostgresqlRepo(dsn string) (*PostgresqlRepo, error) {
	if dsn == "" {
		return nil, errEmptyDSN
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	return &PostgresqlRepo{db: db}, nil
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
