package stats

import (
	"errors"
	"fmt"
	"net"

	"github.com/Skifskii/link-shortener/internal/model"
)

var ErrIPNotAllowed = errors.New("IP not allowed to access stats")

type StatsService struct {
	repo          Repo
	trustedSubnet *net.IPNet
}

type Repo interface {
	GetUsersCount() (int, error)
	GetURLsCount() (int, error)
}

func New(trustedSubnet string, repo Repo) (*StatsService, error) {
	_, ipNet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CIDR: %w", err)
	}

	return &StatsService{trustedSubnet: ipNet, repo: repo}, nil
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

func (s *StatsService) isIPAllowed(ip net.IP) bool {
	if s.trustedSubnet == nil {
		return false
	}
	return s.trustedSubnet.Contains(ip)
}
