package dbping

import "errors"

var errNoDB = errors.New("database not specified")

type pinger interface {
	Ping() error
}

type DBPingService struct {
	repo pinger
}

func New(repo pinger) *DBPingService {
	return &DBPingService{repo: repo}
}

func (d *DBPingService) Ping() error {
	if d.repo == nil {
		return errNoDB
	}

	return d.repo.Ping()
}
