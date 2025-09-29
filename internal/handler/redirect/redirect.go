package redirect

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ShortRedirecter interface {
	Redirect(shortURL string) (longURL string, err error)
}

func New(sr ShortRedirecter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortURL := chi.URLParam(r, "id")
		if shortURL == "" {
			http.Error(w, "id не указан", http.StatusBadRequest)
			return
		}

		longURL, err := sr.Redirect(shortURL)
		if err != nil {
			http.Error(w, "Ссылка не найдена", http.StatusNotFound)
			return
		}

		w.Header().Set("Location", longURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
