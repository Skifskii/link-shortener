package shorten

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Skifskii/link-shortener/internal/middleware/authmw"
	"github.com/Skifskii/link-shortener/internal/model"
	"github.com/Skifskii/link-shortener/internal/repository"
)

type Shortener interface {
	Shorten(userID int, longURL string) (shortURL string, err error)
}

func New(s Shortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// десериализуем запрос в структуру модели
		var req model.Request
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		status := http.StatusCreated

		userID, ok := r.Context().Value(authmw.UserIDKey).(int)
		if !ok {
			http.Error(w, "Ошибка при определении user_id", http.StatusInternalServerError)
			return
		}

		// Сокращаем ссылку
		shortURL, err := s.Shorten(userID, req.URL)
		if err != nil {
			if errors.Is(err, repository.ErrOriginalURLAlreadyExists) {
				status = http.StatusConflict
			} else {
				http.Error(w, "Ошибка при генерации короткой ссылки", http.StatusInternalServerError)
				return
			}
		}

		resp := model.Response{
			Result: shortURL,
		}

		// сериализуем ответ сервера
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)

		enc := json.NewEncoder(w)
		enc.Encode(resp)
	}
}
