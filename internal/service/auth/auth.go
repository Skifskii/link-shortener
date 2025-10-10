package auth

import (
	"errors"

	"github.com/golang-jwt/jwt/v4"
)

var errInvalidToken = errors.New("invalid token")
var errUnexpectedSigningMethod = errors.New("unexpected signing method")

type UserAdder interface {
	CreateUser(username string) (userID int, err error)
}

type AuthService struct {
	repo      UserAdder
	secretKey string
}

func New(repo UserAdder, secretKey string) *AuthService {
	return &AuthService{repo: repo, secretKey: secretKey}
}

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

func (a *AuthService) CreateUser(username string) (jwt string, err error) {
	userID, err := a.repo.CreateUser(username)
	if err != nil {
		return "", err
	}

	return a.BuildJWTString(userID)
}

func (a *AuthService) BuildJWTString(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
	})

	return token.SignedString([]byte(a.secretKey))
}

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
