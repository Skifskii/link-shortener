package shorten

import (
	"encoding/json"
	"net/http"

	"github.com/Skifskii/link-shortener/internal/model"
)

type Shortener interface {
	Shorten(longURL string) (shortURL string, err error)
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

		// Сокращаем ссылку
		shortURL, err := s.Shorten(req.URL)
		if err != nil {
			http.Error(w, "Ошибка при генерации короткой ссылки", http.StatusInternalServerError)
			return
		}

		resp := model.Response{
			Result: shortURL,
		}

		// сериализуем ответ сервера
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		enc := json.NewEncoder(w)
		enc.Encode(resp)
	}
}
