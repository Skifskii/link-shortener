package repository

import "errors"

type Repository interface {
	Save(short, original string) (existingShort string, err error)
	Get(short string) (string, error)
}

var ErrOriginalURLAlreadyExists = errors.New("received link already exists and has a short version")
var ErrShortNotFound = errors.New("short URL not found")
var ErrOriginalNotFound = errors.New("original URL not found")
var ErrLinkDeleted = errors.New("link has been deleted")
