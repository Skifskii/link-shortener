package urlobs

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Skifskii/link-shortener/internal/service/audit"
)

type URLObserver struct {
	url    string
	client *http.Client
}

func New(url string) *URLObserver {
	return &URLObserver{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		url: url,
	}
}

func (u *URLObserver) Update(e *audit.Event) {
	if e == nil {
		return
	}
	u.notifyRemoteService(*e)
}

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
	resp.Body.Close()

	return err
}
