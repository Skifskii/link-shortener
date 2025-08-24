package handler

import (
	"io"
	"net/http"

	"github.com/Skifskii/link-shortener/internal/repository"
	"github.com/Skifskii/link-shortener/internal/service/shortener"
)

func CommonHandler(repo repository.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Не удалось прочитать тело запроса", http.StatusInternalServerError)
				return
			}
			longURL := string(body)

			shortURL := shortener.GenerateShortURL()
			if err := repo.Save(shortURL, string(longURL)); err != nil {
				http.Error(w, "Ошибка при сохранении ссылки", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("http://localhost:8080/" + shortURL))
		case http.MethodGet:
			shortURL := r.URL.Path[1:]
			longURL, err := repo.Get(shortURL)
			if err != nil {
				http.Error(w, "Ссылка не найдена", http.StatusNotFound)
				return
			}

			w.Header().Set("Location", longURL)
			w.WriteHeader(http.StatusTemporaryRedirect)
			w.Write([]byte(longURL))
		default:
			http.Error(w, "Метод не поддерживается", http.StatusBadRequest)
			return
		}
	}
}
