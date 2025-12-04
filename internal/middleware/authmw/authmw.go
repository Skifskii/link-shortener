package authmw

import (
	"context"
	"net/http"
)

type ContextKey string

const UserIDKey ContextKey = "user_id"

type Auther interface {
	CreateUser(username string) (jwt string, err error)
	GetUserID(tokenString string) (int, error)
}

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
