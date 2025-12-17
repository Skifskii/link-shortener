// Package ping предоставляет handler для проверки доступности базы данных.
package ping

import "net/http"

type pinger interface {
	Ping() error
}

// New возвращает HTTP-хендлер для endpoint /ping, который проверяет Ping репозитория.
func New(p pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if err := p.Ping(); err != nil {
			http.Error(w, "Ошибка проверки подключения к БД", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
