package stats

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

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

		ip, err := resolveIP(r)
		if err != nil {
			http.Error(w, "Ошибка при определении IP-адреса", http.StatusInternalServerError)
			return
		}

		// пытаемся получить статистику
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

func resolveIP(r *http.Request) (net.IP, error) {
	addr := r.RemoteAddr
	// метод возвращает адрес в формате host:port
	// нужна только подстрока host
	ipStr, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	// парсим ip
	ip := net.ParseIP(ipStr)
	if ip != nil {
		return ip, nil
	}

	// смотрим заголовок запроса X-Real-IP
	ipStr = r.Header.Get("X-Real-IP")
	// парсим ip
	ip = net.ParseIP(ipStr)
	if ip == nil {
		// если заголовок X-Real-IP пуст, пробуем X-Forwarded-For
		// этот заголовок содержит адреса отправителя и промежуточных прокси
		// в виде 203.0.113.195, 70.41.3.18, 150.172.238.178
		ips := r.Header.Get("X-Forwarded-For")
		// разделяем цепочку адресов
		ipStrs := strings.Split(ips, ",")
		// интересует только первый
		ipStr = ipStrs[0]
		// парсим
		ip = net.ParseIP(ipStr)
	}
	if ip == nil {
		return nil, fmt.Errorf("failed parse ip from http header")
	}
	return ip, nil
}
