package shortener

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

type URLSaveGetter interface {
	Save(shortURL, longURL string) error
	Get(shortURL string) (string, error)
}

type ShorterService struct {
	baseURL string
	length  int
	repo    URLSaveGetter
}

func New(baseURL string, length int, repo URLSaveGetter) *ShorterService {
	return &ShorterService{baseURL: baseURL, length: length, repo: repo}
}

func (s *ShorterService) Shorten(longURL string) (shortURL string, err error) {
	shortCode, err := s.generateShortCode()
	if err != nil {
		return "", err
	}

	shortURL = s.baseURL + "/" + shortCode

	if err := s.repo.Save(shortURL, longURL); err != nil {
		return "", err
	}

	return shortURL, nil
}

func (s *ShorterService) Redirect(shortURL string) (longURL string, err error) {
	longURL, err = s.repo.Get(s.baseURL + "/" + shortURL)
	if err != nil {
		return "", fmt.Errorf("ссылка не найдена: %w", err)
	}

	// защита от пустой строки в хранилище
	if strings.TrimSpace(longURL) == "" {
		return "", errors.New("longURL является пустой строкой")
	}

	// Если в longURL нет схемы (http:// или https://), добавить http://
	if !strings.Contains(longURL, "://") {
		longURL = "http://" + longURL
	}

	return longURL, err
}

func (s *ShorterService) generateShortCode() (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, s.length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		b[i] = letters[n.Int64()]
	}

	return string(b), nil
}
