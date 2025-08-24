package inmemory

import (
	"sync"

	"github.com/Skifskii/link-shortener/internal/repository"
)

var _ repository.Repository = (*InMemoryRepo)(nil)

type InMemoryRepo struct {
	store map[string]string
	mu    sync.Mutex
}

func New() *InMemoryRepo {
	return &InMemoryRepo{
		store: make(map[string]string),
	}
}

func (r *InMemoryRepo) Save(short, original string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[short] = original
	return nil
}

func (r *InMemoryRepo) Get(short string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	original, exists := r.store[short]
	if !exists {
		return "", nil
	}
	return original, nil
}
