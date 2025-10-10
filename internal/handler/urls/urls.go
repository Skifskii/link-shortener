package urls

import (
	"fmt"
	"net/http"

	"github.com/Skifskii/link-shortener/internal/middleware/authmw"
)

func New() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		userID, ok := r.Context().Value(authmw.UserIDKey).(int)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// TODO: возвращать все ссылки пользователя
		fmt.Println(userID)
	}
}
