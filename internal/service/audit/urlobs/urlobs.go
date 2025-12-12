// Package urlobs реализует наблюдатель аудита, который отправляет события на удалённый URL.
package urlobs

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Skifskii/link-shortener/internal/service/audit"
)

// URLObserver отправляет JSON-представление события на указанный URL по HTTP.
type URLObserver struct {
	url    string
	client *http.Client
}

// New создаёт новый URLObserver с таймаутом клиента 10 секунд.
func New(url string) *URLObserver {
	return &URLObserver{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		url: url,
	}
}

// Update отправляет событие на удалённый сервис (не блокирующая операция).
func (u *URLObserver) Update(e *audit.Event) {
	if e == nil {
		return
	}
	u.notifyRemoteService(*e)
}

// notifyRemoteService сериализует событие и выполняет POST-запрос к удалённому URL.
func (u *URLObserver) notifyRemoteService(e audit.Event) error {
	body, err := json.Marshal(e)

	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", u.url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := u.client.Do(req)
	if resp != nil {
		resp.Body.Close()
	}

	return err
}
