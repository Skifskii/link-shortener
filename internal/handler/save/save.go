package save

import (
	"io"
	"net/http"
)

type ShortGenerator interface {
	GenerateShort() (string, error)
}

type URLSaver interface {
	Save(shortURL, longURL string) error
}

func New(us URLSaver, s ShortGenerator, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Не удалось прочитать тело запроса", http.StatusInternalServerError)
			return
		}
		longURL := string(body)
		r.Body.Close()

		shortURL, err := s.GenerateShort()
		if err != nil {
			http.Error(w, "Ошибка при генерации короткой ссылки", http.StatusInternalServerError)
			return
		}

		if err := us.Save(shortURL, string(longURL)); err != nil {
			http.Error(w, "Ошибка при сохранении ссылки", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(baseURL + "/" + shortURL))
	}
}
