// Package auth реализует простую JWT-аутентификацию и создание пользователей.
package auth

import (
	"errors"

	"github.com/golang-jwt/jwt/v4"
)

var errInvalidToken = errors.New("invalid token")
var errUnexpectedSigningMethod = errors.New("unexpected signing method")

// UserAdder интерфейс для создания пользователя в хранилище.
type UserAdder interface {
	CreateUser(username string) (userID int, err error)
}

// AuthService отвечает за создание JWT и извлечение userID из токена.
type AuthService struct {
	repo      UserAdder
	secretKey string
}

// New создаёт новый AuthService с указанным хранилищем и секретным ключом.
func New(repo UserAdder, secretKey string) *AuthService {
	return &AuthService{repo: repo, secretKey: secretKey}
}

// Claims используется для хранения пользовательских данных в JWT.
type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

// CreateUser создаёт пользователя через репозиторий и возвращает JWT строку.
func (a *AuthService) CreateUser(username string) (jwt string, err error) {
	userID, err := a.repo.CreateUser(username)
	if err != nil {
		return "", err
	}

	return a.BuildJWTString(userID)
}

// BuildJWTString формирует JWT строку для указанного userID.
func (a *AuthService) BuildJWTString(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
	})

	return token.SignedString([]byte(a.secretKey))
}

// GetUserID парсит JWT и возвращает userID, либо ошибку.
func (a *AuthService) GetUserID(tokenString string) (int, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errUnexpectedSigningMethod
			}
			return []byte(a.secretKey), nil
		},
	)
	if err != nil {
		return -1, err
	}

	if !token.Valid {
		return -1, errInvalidToken
	}

	return claims.UserID, nil
}
