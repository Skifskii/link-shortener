// Package authmw предоставляет middleware для извлечения/создания JWT-cookie и
// добавления user_id в контекст запроса.
package authmw

import (
	"context"
	"net/http"
)

type ContextKey string

const UserIDKey ContextKey = "user_id"

// Auther - интерфейс для работы с пользователями и JWT-токенами.
type Auther interface {
	// CreateUser создает нового пользователя и возвращает его JWT-токен.
	CreateUser(username string) (jwt string, err error)
	// GetUserID извлекает идентификатор пользователя из JWT-токена.
	GetUserID(tokenString string) (int, error)
}

// AuthMiddleware проверяет наличие cookie "jwt" и помещает user_id в контекст запроса.
func AuthMiddleware(a Auther) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, _ := r.Cookie("jwt")

			var token string
			var err error

			// Если куки нет или кука не проходит проверку подлинности - записываем ее
			if cookie == nil {
				// TODO: добавить проверку подлинности
				token, err = a.CreateUser("")
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				http.SetCookie(w, &http.Cookie{
					Name:  "jwt",
					Value: token,
				})
			} else {
				token = cookie.Value
			}

			userID, err := a.GetUserID(token)
			if err != nil {
				// TODO: сделать обработку подробнее
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
