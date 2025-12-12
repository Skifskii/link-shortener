package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/Skifskii/link-shortener/internal/handler/api/shorten"
	batchhandler "github.com/Skifskii/link-shortener/internal/handler/api/shorten/batch"
	urls "github.com/Skifskii/link-shortener/internal/handler/api/user/urls"
	"github.com/Skifskii/link-shortener/internal/handler/ping"
	"github.com/Skifskii/link-shortener/internal/handler/redirect"
	savehandler "github.com/Skifskii/link-shortener/internal/handler/save"
	"github.com/Skifskii/link-shortener/internal/middleware/authmw"
	"github.com/Skifskii/link-shortener/internal/model"
	"github.com/Skifskii/link-shortener/internal/service/audit"
	"github.com/go-chi/chi/v5"
)

// фиксированная реализация короткого сервиса, используемая в примерах
type fixedShortener struct {
	base string
}

func (f *fixedShortener) Shorten(userID int, longURL string) (string, error) {
	return f.base + "/abc", nil
}

func (f *fixedShortener) Redirect(shortURL string) (string, error) {
	// возвращаем детерминированный "длинный" URL
	return "https://example.com/original/" + shortURL, nil
}

func (f *fixedShortener) BatchShorten(userID int, reqBatch []model.RequestArrayElement) ([]model.ResponseArrayElement, error) {
	resp := make([]model.ResponseArrayElement, 0, len(reqBatch))
	for i := range reqBatch {
		resp = append(resp, model.ResponseArrayElement{
			CorrelationID: reqBatch[i].CorrelationID,
			ShortURL:      f.base + "/batch/" + reqBatch[i].CorrelationID,
		})
	}
	return resp, nil
}

func (f *fixedShortener) GetUserPairs(userID int) ([]model.ResponsePairElement, error) {
	return []model.ResponsePairElement{{ShortURL: f.base + "/1", OriginalURL: "https://orig.example/1"}}, nil
}

func (f *fixedShortener) DeleteUserLinks(userID int, shortURLs []string) error {
	return nil
}

// noop-аудит
type noopAudit struct{}

func (n *noopAudit) NotifyAll(ev *audit.Event) {}

// добавляем реализацию pinger на уровне пакета
type pingerImpl struct{}

func (p *pingerImpl) Ping() error { return nil }

// Example: / (POST) — сохранение одиночной ссылки
func ExampleNew_save() {
	s := &fixedShortener{base: "http://short"}
	h := savehandler.New(s, &noopAudit{})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://original.example/path"))
	// в реальном приложении user_id добавляется middleware; для примера явно укажем
	req = req.WithContext(context.WithValue(req.Context(), authmw.UserIDKey, 1))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(strings.TrimSpace(rr.Body.String()))

	// Output:
	// 201
	// http://short/abc
}

// Example: GET /{id} — редирект
func ExampleRedirect() {
	s := &fixedShortener{base: "http://short"}
	h := redirect.New(s, &noopAudit{})
	// h := func() http.Handler {
	// 	// используем тот же handler из redirect.New, но импортать пакет redirect здесь
	// 	// чтобы не дублировать зависимости, создадим минимальный wrapper, который
	// 	// использует Redirect метод напрямую.
	// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 		id := chi.URLParam(r, "id")
	// 		if id == "" {
	// 			w.WriteHeader(http.StatusBadRequest)
	// 			return
	// 		}
	// 		long, err := s.Redirect(id)
	// 		if err != nil {
	// 			w.WriteHeader(http.StatusNotFound)
	// 			return
	// 		}
	// 		w.Header().Set("Location", long)
	// 		w.WriteHeader(http.StatusTemporaryRedirect)
	// 	})
	// }()

	req := httptest.NewRequest(http.MethodGet, "/abc", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	// добавить user_id, т.к. в оригинальном handler-е он используется для аудита
	req = req.WithContext(context.WithValue(req.Context(), authmw.UserIDKey, 1))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Header().Get("Location"))

	// Output:
	// 307
	// https://example.com/original/abc
}

// Example: POST /api/shorten (JSON)
func ExampleNew_shorten() {
	s := &fixedShortener{base: "http://api"}
	h := shorten.New(s, &noopAudit{})

	reqBody := model.Request{URL: "https://orig.example"}
	b, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(b))
	req = req.WithContext(context.WithValue(req.Context(), authmw.UserIDKey, 1))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(strings.TrimSpace(rr.Body.String()))

	// Output:
	// 201
	// {"result":"http://api/abc"}
}

// Example: POST /api/shorten/batch
func ExampleNew_batchhandler() {
	s := &fixedShortener{base: "http://api"}
	h := batchhandler.New(s)

	reqArray := []model.RequestArrayElement{{CorrelationID: "c1", OriginalURL: "https://a"}}
	b, _ := json.Marshal(reqArray)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(b))
	req = req.WithContext(context.WithValue(req.Context(), authmw.UserIDKey, 1))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(strings.TrimSpace(rr.Body.String()))

	// Output:
	// 201
	// [{"correlation_id":"c1","short_url":"http://api/batch/c1"}]
}

// Example: GET /api/user/urls
func ExampleNew_urls() {
	s := &fixedShortener{base: "http://api"}
	h := urls.New(s)

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req = req.WithContext(context.WithValue(req.Context(), authmw.UserIDKey, 1))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(strings.TrimSpace(rr.Body.String()))

	// Output:
	// 200
	// [{"short_url":"http://api/1","original_url":"https://orig.example/1"}]
}

// Example: DELETE /api/user/urls
func ExampleNewDelete_urls() {
	s := &fixedShortener{base: "http://api"}
	h := urls.NewDelete(s)

	b, _ := json.Marshal([]string{"http://api/1"})
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(b))
	req = req.WithContext(context.WithValue(req.Context(), authmw.UserIDKey, 1))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)

	// Output:
	// 202
}

// Example: GET /ping
func ExampleNew_ping() {
	pr := &pingerImpl{}
	h := ping.New(pr)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)

	// Output:
	// 200
}
