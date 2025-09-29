package dbping

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
	return d.repo.Ping()
}
