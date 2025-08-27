package shortener

import "math/rand"

type Shortener struct {
	length int
}

func New(length int) Shortener {
	return Shortener{length: length}
}

func (s Shortener) GenerateShort() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, s.length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
