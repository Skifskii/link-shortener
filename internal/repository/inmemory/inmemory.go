// Package inmemory предоставляет тестовое/временное хранилище ссылок в памяти.
package inmemory

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Skifskii/link-shortener/internal/model"
	"github.com/Skifskii/link-shortener/internal/repository"
)

type originalLink struct {
	link    string
	deleted bool
}

type user struct {
	store map[string]originalLink
}

// InMemoryRepo - структура для хранения пользователей и их ссылок в памяти.
type InMemoryRepo struct {
	users map[int]user
	mu    sync.Mutex
}

// New создаёт новый InMemoryRepo.
func New() *InMemoryRepo {
	return &InMemoryRepo{
		users: make(map[int]user),
	}
}

// Save сохраняет пару short->original для указанного userID.
// Если original уже присутствует в хранилище, возвращается сохранённый short и
// ошибка repository.ErrOriginalURLAlreadyExists.
func (r *InMemoryRepo) Save(userID int, short, original string) (savedShort string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[userID]; !ok {
		r.users[userID] = user{make(map[string]originalLink)}
	}

	// Проверяем, есть ли original среди сохранённых
	for s, o := range r.users[userID].store {
		if o.link == original {
			return s, repository.ErrOriginalURLAlreadyExists
		}
	}

	r.users[userID].store[short] = originalLink{link: original}
	return "", nil
}

// SaveBatch сохраняет пакет ссылок.
func (r *InMemoryRepo) SaveBatch(_ int, shortURLs, longURLs []string) error {
	for i, short := range shortURLs {
		if _, err := r.Save(0, short, longURLs[i]); err != nil { // TODO:
			return err
		}
	}
	return nil
}

// Get возвращает оригинальный URL по короткой ссылке.
func (r *InMemoryRepo) Get(short string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var original originalLink
	var exists bool

	for _, u := range r.users {
		original, exists = u.store[short]
		if exists {
			if original.deleted {
				return "", repository.ErrLinkDeleted
			}
			return original.link, nil
		}
	}

	// TODO: если original.deleted, то вернуть ошибку

	return "", repository.ErrShortNotFound
}

// GetUserPairs возвращает список пар short->original для пользователя.
func (r *InMemoryRepo) GetUserPairs(userID int) ([]model.ResponsePairElement, error) {
	return []model.ResponsePairElement{}, nil
}

// CreateUser создаёт нового пользователя и возвращает его ID.
func (r *InMemoryRepo) CreateUser(username string) (userID int, err error) {
	for i := 1; i < 100_000; i++ {
		if _, ok := r.users[i]; !ok {
			r.users[i] = user{make(map[string]originalLink)}
			return i, nil
		}
	}
	return -1, errors.New("failed creating user")
}

// DeleteBatchOfLinks помечает указанные ссылки пользователя как удалённые.
func (r *InMemoryRepo) DeleteBatchOfLinks(userID int, shortURLs []string) error {
	for _, short := range shortURLs {
		if err := r.deleteLinkByShort(userID, short); err != nil { // TODO: здесь не передается baseURL. Нужно изменить логику Save - хранить только сокращенную ссылку, и добавить в ShortenerService функцию добавления baseURL.
			return err
		}
	}
	return nil
}

func (r *InMemoryRepo) deleteLinkByShort(userID int, shortURL string) error {
	user, ok := r.users[userID]
	if !ok {
		return fmt.Errorf("can't find user with user_id=%d", userID)
	}

	originalURL, ok := user.store[shortURL]
	if !ok {
		return fmt.Errorf("can't find shortURL=%s in user (user_id=%d) list", shortURL, userID)
	}

	originalURL.deleted = true
	user.store[shortURL] = originalURL

	return nil
}
