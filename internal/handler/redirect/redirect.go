package redirect

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type URLGetter interface {
	Get(shortURL string) (string, error)
}

func New(ug URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortURL := chi.URLParam(r, "id")

		longURL, err := ug.Get(shortURL)
		if err != nil {
			http.Error(w, "Ссылка не найдена", http.StatusNotFound)
			return
		}

		w.Header().Set("Location", longURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
