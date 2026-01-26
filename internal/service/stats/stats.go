package stats

import (
	"errors"
	"net"

	"github.com/Skifskii/link-shortener/internal/model"
)

var ErrIPNotAllowed = errors.New("IP not allowed to access stats")

type StatsService struct {
	repo          Repo
	trustedSubnet string
}

type Repo interface {
	GetUsersCount() (int, error)
	GetURLsCount() (int, error)
}

func New(trustedSubnet string, repo Repo) *StatsService {
	return &StatsService{trustedSubnet: trustedSubnet, repo: repo}
}

func (s *StatsService) GetIfAllowed(ip net.IP) (model.StatsResponse, error) {
	// Если IP-адрес не в доверенной подсети, возвращаем ошибку
	if !s.isIPAllowed(ip) {
		return model.StatsResponse{}, ErrIPNotAllowed
	}

	// Получаем статистику из репозитория
	usersCount, err := s.repo.GetUsersCount()
	if err != nil {
		return model.StatsResponse{}, err
	}
	urlsCount, err := s.repo.GetURLsCount()
	if err != nil {
		return model.StatsResponse{}, err
	}

	return model.StatsResponse{
		URLs:  urlsCount,
		Users: usersCount,
	}, nil
}

func (_ *StatsService) isIPAllowed(_ net.IP) bool {
	return true // TODO: описать логику проверки
}
