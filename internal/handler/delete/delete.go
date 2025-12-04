package delete

import (
	"encoding/json"
	"net/http"

	"github.com/Skifskii/link-shortener/internal/middleware/authmw"
)

type Shortener interface {
	DeleteUserLinks(userID int, shortURLs []string) error
}

func New(s Shortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		userID, ok := r.Context().Value(authmw.UserIDKey).(int)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// десериализуем запрос в структуру модели
		var shortURLs []string
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&shortURLs); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := s.DeleteUserLinks(userID, shortURLs); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// сериализуем ответ сервера
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
	}
}
