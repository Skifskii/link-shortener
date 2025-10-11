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

type InMemoryRepo struct {
	users map[int]user
	mu    sync.Mutex
}

func New() *InMemoryRepo {
	return &InMemoryRepo{
		users: make(map[int]user),
	}
}

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

func (r *InMemoryRepo) SaveBatch(shortURLs, longURLs []string) error {
	for i, short := range shortURLs {
		if _, err := r.Save(0, short, longURLs[i]); err != nil { // TODO:
			return err
		}
	}
	return nil
}

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

func (r *InMemoryRepo) GetUserPairs(userID int) ([]model.ResponsePairElement, error) {
	return []model.ResponsePairElement{}, nil
}

func (r *InMemoryRepo) CreateUser(username string) (userID int, err error) {
	for i := 1; i < 100_000; i++ {
		if _, ok := r.users[i]; !ok {
			r.users[i] = user{make(map[string]originalLink)}
			return i, nil
		}
	}
	return -1, errors.New("failed creating user")
}

func (r *InMemoryRepo) DeleteLinkByShort(userID int, shortURL string) error {
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
