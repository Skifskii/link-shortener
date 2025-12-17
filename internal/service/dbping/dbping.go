// Package dbping предоставляет сервис проверки доступности базы данных.
package dbping

import "errors"

var errNoDB = errors.New("database not specified")

// DBPingService реализует проверку доступности базы данных через метод Ping.
type DBPingService struct {
	repo pinger
}

// New создаёт DBPingService с указанным репозиторием (реализует Ping).
func New(repo pinger) *DBPingService {
	return &DBPingService{repo: repo}
}

// Ping вызывает Ping на репозитории и возвращает ошибку при отсутствии репозитория.
func (d *DBPingService) Ping() error {
	if d.repo == nil {
		return errNoDB
	}

	return d.repo.Ping()
}

type pinger interface {
	Ping() error
}
