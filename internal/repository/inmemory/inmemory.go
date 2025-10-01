package inmemory

import (
	"sync"

	"github.com/Skifskii/link-shortener/internal/repository"
)

var _ repository.Repository = (*InMemoryRepo)(nil)

type InMemoryRepo struct {
	Store map[string]string
	mu    sync.Mutex
}

func New() *InMemoryRepo {
	return &InMemoryRepo{
		Store: make(map[string]string),
	}
}

func (r *InMemoryRepo) Save(short, original string) (savedShort string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем, есть ли original среди сохранённых
	for s, o := range r.Store {
		if o == original {
			return s, repository.ErrOriginalURLAlreadyExists
		}
	}

	r.Store[short] = original
	return "", nil
}

func (r *InMemoryRepo) SaveBatch(shortURLs, longURLs []string) error {
	for i, short := range shortURLs {
		if _, err := r.Save(short, longURLs[i]); err != nil { // TODO:
			return err
		}
	}
	return nil
}

func (r *InMemoryRepo) Get(short string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	original, exists := r.Store[short]
	if !exists {
		return "", repository.ErrShortNotFound
	}
	return original, nil
}
