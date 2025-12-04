package save

import (
	"errors"
	"io"
	"net/http"

	"github.com/Skifskii/link-shortener/internal/middleware/authmw"
	"github.com/Skifskii/link-shortener/internal/repository"
)

type Shortener interface {
	Shorten(userID int, longURL string) (shortURL string, err error)
}

func New(s Shortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Не удалось прочитать тело запроса", http.StatusInternalServerError)
			return
		}

		longURL := string(body)
		r.Body.Close()

		userID, ok := r.Context().Value(authmw.UserIDKey).(int)
		if !ok {
			http.Error(w, "Ошибка при определении user_id", http.StatusInternalServerError)
			return
		}

		shortURL, err := s.Shorten(userID, longURL)
		if err != nil {
			if errors.Is(err, repository.ErrOriginalURLAlreadyExists) {
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(shortURL))
				return
			}
			http.Error(w, "Ошибка при генерации короткой ссылки", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	}
}
