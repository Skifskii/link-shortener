package shortener

import (
	"crypto/rand"
	"math/big"
)

type Shortener struct {
	length int
}

func New(length int) Shortener {
	return Shortener{length: length}
}

func (s Shortener) GenerateShort() (string, error) {
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
