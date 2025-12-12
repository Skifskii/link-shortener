// Package shortener реализует логику сокращения ссылок: генерацию кодов, создание
// укороченных ссылок и работу с пакетными операциями.
package shortener

import (
	"crypto/rand"
	"errors"
	"strings"

	"github.com/Skifskii/link-shortener/internal/model"
)

// URLSaveGetter - интерфейс для сохранения и получения URL из репозитория.
type URLSaveGetter interface {
	Save(userID int, shortURL, longURL string) (existingShort string, err error)
	Get(shortURL string) (string, error)
	SaveBatch(userID int, shortURLs, longURLs []string) error
	GetUserPairs(userID int) ([]model.ResponsePairElement, error)
	DeleteBatchOfLinks(userID int, shortURL []string) error
}

// ShorterService предоставляет методы для сокращения ссылок и работы с хранилищем.
type ShorterService struct {
	baseURL string
	length  int
	repo    URLSaveGetter
}

// New создаёт новый экземпляр сервиса сокращения ссылок.
func New(baseURL string, length int, repo URLSaveGetter) *ShorterService {
	return &ShorterService{baseURL: baseURL, length: length, repo: repo}
}

// Shorten создаёт короткую ссылку для переданного longURL и сохраняет её в репозитории.
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

// BatchShorten обрабатывает массив запросов и создаёт массив ответов с короткими ссылками.
func (s *ShorterService) BatchShorten(userID int, reqBatch []model.RequestArrayElement) (respBatch []model.ResponseArrayElement, err error) {
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

	if err := s.repo.SaveBatch(userID, shortURLs, longURLs); err != nil {
		return nil, err
	}

	return respBatch, nil
}

// Redirect возвращает исходный URL по краткой части shortURL и добавляет схему, если нужно.
func (s *ShorterService) Redirect(shortURL string) (longURL string, err error) {
	longURL, err = s.repo.Get(s.baseURL + "/" + shortURL)
	if err != nil {
		return "", err
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

// generateShortCode генерирует короткий код заданной длины.
func (s *ShorterService) generateShortCode() (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if s.length <= 0 {
		return "", nil
	}

	letterBytes := []byte(letters)
	l := byte(len(letterBytes)) // 62
	// порог для rejection sampling: 256 - (256 % 62) = 248
	const threshold = 256 - (256 % 62)

	result := make([]byte, s.length)
	// буфер немного больше, чтобы реже вызывать rand.Read
	bufSize := s.length + s.length/2
	if bufSize < 16 {
		bufSize = 16
	}
	buf := make([]byte, bufSize)

	i := 0
	for i < s.length {
		_, err := rand.Read(buf)
		if err != nil {
			return "", err
		}
		for _, b := range buf {
			if int(b) >= threshold {
				// отбрасываем, чтобы избежать биаса
				continue
			}
			result[i] = letterBytes[b%l]
			i++
			if i == s.length {
				break
			}
		}
	}

	return string(result), nil
}

// GetUserPairs возвращает пары short->original для заданного пользователя.
func (s *ShorterService) GetUserPairs(userID int) ([]model.ResponsePairElement, error) {
	return s.repo.GetUserPairs(userID)
}

// DeleteUserLinks помечает набор коротких ссылок пользователя как удалённые.
func (s *ShorterService) DeleteUserLinks(userID int, shortURLs []string) error {
	for i := 0; i < len(shortURLs); i++ {
		shortURLs[i] = s.baseURL + "/" + shortURLs[i]
	}

	return s.repo.DeleteBatchOfLinks(userID, shortURLs)
}
