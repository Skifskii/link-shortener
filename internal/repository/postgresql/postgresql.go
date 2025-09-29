package postgresql

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresqlRepo struct {
	db *sql.DB
}

func NewPostgresqlRepo(dsn string) (*PostgresqlRepo, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	return &PostgresqlRepo{db: db}, nil
}

// func (pr *PostgresqlRepo) Save(short, original string) error {
// 	return nil // TODO:
// }

// func (pr *PostgresqlRepo) Get(short string) (original string, err error) {
// 	return "", nil // TODO:
// }

func (pr *PostgresqlRepo) Ping() error {
	return pr.db.Ping()
}

func (pr *PostgresqlRepo) Close() error {
	return pr.db.Close()
}
