package redirect

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type URLGetter interface {
	Get(shortURL string) (string, error)
}

func New(ug URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortURL := chi.URLParam(r, "id")
		if shortURL == "" {
			http.Error(w, "id не указан", http.StatusBadRequest)
			return
		}

		longURL, err := ug.Get(shortURL)
		if err != nil {
			http.Error(w, "Ссылка не найдена", http.StatusNotFound)
			return
		}

		// защита от пустой строки в хранилище
		if strings.TrimSpace(longURL) == "" {
			http.Error(w, "Ссылка не найдена", http.StatusNotFound)
			return
		}

		// Если в longURL нет схемы (http:// или https://), добавить http://
		if !strings.Contains(longURL, "://") {
			longURL = "http://" + longURL
		}

		w.Header().Set("Location", longURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
