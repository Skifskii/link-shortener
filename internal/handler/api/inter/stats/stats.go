package stats

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"

	"github.com/Skifskii/link-shortener/internal/model"
	"github.com/Skifskii/link-shortener/internal/service/stats"
)

// StatsGetter - интерфейс для получения статистики.
type StatsGetter interface {
	GetIfAllowed(ip net.IP) (model.StatsResponse, error)
}

// New возвращает HTTP-хендлер для endpoint /api/shorten.
func New(sg StatsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// пытаемся получить статистику
		ip := net.ParseIP(r.RemoteAddr) // TODO: получить IP клиента (заглянуть в X-Forwarded-For)
		resp, err := sg.GetIfAllowed(ip)
		if err != nil {
			if errors.Is(err, stats.ErrIPNotAllowed) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			} else {
				http.Error(w, "Ошибка при генерации короткой ссылки", http.StatusInternalServerError)
				return
			}
		}

		// сериализуем ответ сервера
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		enc := json.NewEncoder(w)
		enc.Encode(resp)
	}
}
