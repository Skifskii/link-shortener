package shortener

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/Skifskii/link-shortener/internal/model"
)

type URLSaveGetter interface {
	Save(userID int, shortURL, longURL string) (existingShort string, err error)
	Get(shortURL string) (string, error)
	SaveBatch(shortURLs, longURLs []string) error
	GetUserPairs(userID int) ([]model.ResponsePairElement, error)
}

type ShorterService struct {
	baseURL string
	length  int
	repo    URLSaveGetter
}

func New(baseURL string, length int, repo URLSaveGetter) *ShorterService {
	return &ShorterService{baseURL: baseURL, length: length, repo: repo}
}

func (s *ShorterService) Shorten(userID int, longURL string) (shortURL string, err error) {
	shortCode, err := s.generateShortCode()
	if err != nil {
		return "", err
	}

	shortURL = s.baseURL + "/" + shortCode

	if existingShort, err := s.repo.Save(userID, shortURL, longURL); err != nil {
		return existingShort, err
	}

	return shortURL, nil
}

func (s *ShorterService) BatchShorten(reqBatch []model.RequestArrayElement) (respBatch []model.ResponseArrayElement, err error) {
	respBatch = make([]model.ResponseArrayElement, 0, len(reqBatch))

	longURLs := make([]string, 0, len(reqBatch))
	shortURLs := make([]string, 0, len(reqBatch))

	for _, req := range reqBatch {
		longURLs = append(longURLs, req.OriginalURL)
		shortCode, err := s.generateShortCode()
		if err != nil {
			return nil, err
		}

		shortURL := s.baseURL + "/" + shortCode
		shortURLs = append(shortURLs, shortURL)

		respBatch = append(respBatch, model.ResponseArrayElement{
			CorrelationID: req.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	if err := s.repo.SaveBatch(shortURLs, longURLs); err != nil {
		return nil, err
	}

	return respBatch, nil
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

func (s *ShorterService) GetUserPairs(userID int) ([]model.ResponsePairElement, error) {
	return s.repo.GetUserPairs(userID)
}
