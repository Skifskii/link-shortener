// Package postgresql реализует репозиторий на основе PostgreSQL с поддержкой
// миграций и фоновой обработки удаления ссылок.
package postgresql

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

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

// PostgresqlRepo реализует хранение ссылок в PostgreSQL.
type PostgresqlRepo struct {
	db          *sql.DB
	delTaskChan chan deleteTask
}

type deleteTask struct {
	userID   int
	shortURL string
}

// NewPostgresqlRepo создаёт подключение к PostgreSQL, запускает миграции и фонового работника удаления.
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

	repo := &PostgresqlRepo{
		db:          db,
		delTaskChan: make(chan deleteTask, 100),
	}

	go repo.deleteWorker()

	return repo, nil
}

func (pr *PostgresqlRepo) deleteWorker() {
	ticker := time.NewTicker(10 * time.Second)

	var tasks []deleteTask

	for {
		select {
		case dt := <-pr.delTaskChan:
			tasks = append(tasks, dt)
		case <-ticker.C:
			if len(tasks) == 0 {
				continue
			}
			pr.completeDeleteTasks(tasks)
			tasks = nil
		}
	}
}

func (pr *PostgresqlRepo) completeDeleteTasks(tasks []deleteTask) {
	if len(tasks) == 0 {
		return
	}

	tx, err := pr.db.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		`UPDATE links
			SET is_deleted = TRUE
			WHERE id = (
				SELECT l.id
				FROM links AS l
				JOIN users_links AS ul ON ul.link_id = l.id
				WHERE l.short = $1 AND ul.user_id = $2
				LIMIT 1
			);
	`)
	if err != nil {
		fmt.Printf("tx.Prepare error")
		return
	}
	defer stmt.Close()

	for _, task := range tasks {
		_, err := stmt.Exec(task.shortURL, task.userID)
		if err != nil {
			fmt.Printf("stmt.Exec error")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		fmt.Printf("tx.Commit error")
	}
}

// DeleteBatchOfLinks ставит задачи на удаление ссылок в очередь для фоновой обработки.
func (pr *PostgresqlRepo) DeleteBatchOfLinks(userID int, shortURLs []string) error {
	go func() {
		for _, shortURL := range shortURLs {
			pr.delTaskChan <- deleteTask{userID, shortURL}
		}
	}()

	return nil
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

// Save сохраняет короткую ссылку и привязывает её к пользователю. Возвращает существующий
// short если original уже был сохранён.
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
		VALUES ($1, $2)
		ON CONFLICT (user_id, link_id) DO NOTHING`,
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

// SaveBatch сохраняет пакет ссылок в транзакции.
func (pr *PostgresqlRepo) SaveBatch(userID int, shortURLs, longURLs []string) error {
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

	stmtLinks, err := tx.Prepare(`
		INSERT INTO links (short, original)
		VALUES ($1, $2)
		ON CONFLICT (original) DO UPDATE
			SET short = links.short
		RETURNING id;
	`)
	if err != nil {
		return err
	}
	defer stmtLinks.Close()

	stmtUsersLinks, err := tx.Prepare(`
		INSERT INTO users_links (user_id, link_id)
			VALUES ($1, $2)
			ON CONFLICT (user_id, link_id) DO NOTHING;
	`)
	if err != nil {
		return err
	}
	defer stmtLinks.Close()

	for i, short := range shortURLs {
		var linkID int
		err := stmtLinks.QueryRow(short, longURLs[i]).Scan(&linkID)
		if err != nil {
			return err
		}

		_, err = stmtUsersLinks.Exec(userID, linkID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Get возвращает оригинальный URL по короткой ссылке, проверяя флаг удаления.
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

// Ping проверяет доступность подключения к базе данных.
func (pr *PostgresqlRepo) Ping() error {
	return pr.db.Ping()
}

// Close закрывает соединение с базой данных.
func (pr *PostgresqlRepo) Close() error {
	return pr.db.Close()
}

// CreateUser создаёт нового пользователя в базе данных и возвращает его ID.
func (pr *PostgresqlRepo) CreateUser(username string) (userID int, err error) {
	err = pr.db.QueryRow(
		`INSERT INTO users (username)
		VALUES ($1)
		RETURNING id`,
		username,
	).Scan(&userID)

	return userID, err
}

// GetUserPairs возвращает все пары short->original, которые принадлежат пользователю.
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

func (pr *PostgresqlRepo) GetUsersCount() (count int, err error) {
	row := pr.db.QueryRow(
		"SELECT COUNT(*) FROM users",
	)
	err = row.Scan(&count)
	return count, err
}

func (pr *PostgresqlRepo) GetURLsCount() (count int, err error) {
	row := pr.db.QueryRow(
		"SELECT COUNT(*) FROM links",
	)
	err = row.Scan(&count)
	return count, err
}
