package inmemory

import (
	"errors"
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

func (r *InMemoryRepo) Save(short, original string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Store[short] = original
	return nil
}

func (r *InMemoryRepo) Get(short string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	original, exists := r.Store[short]
	if !exists {
		return "", errors.New("not found")
	}
	return original, nil
}
