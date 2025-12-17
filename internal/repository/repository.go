package repository

import "errors"

// Repository описывает минимальный набор операций, поддерживаемых репозиторием.
type Repository interface {
	Save(short, original string) (existingShort string, err error)
	Get(short string) (string, error)
}

// Стандартные ошибки репозитория, используемые в разных реализациях.
var (
	ErrOriginalURLAlreadyExists = errors.New("received link already exists and has a short version")
	ErrShortNotFound            = errors.New("short URL not found")
	ErrOriginalNotFound         = errors.New("original URL not found")
	ErrLinkDeleted              = errors.New("link has been deleted")
)
