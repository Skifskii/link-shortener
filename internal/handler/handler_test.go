package handler

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockRepo implements repository.Repository for testing
type mockRepo struct {
	saveErr error
	getErr  error
	data    map[string]string
}

func (m *mockRepo) Save(shortURL, longURL string) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	if m.data == nil {
		m.data = make(map[string]string)
	}
	m.data[shortURL] = longURL
	return nil
}

func (m *mockRepo) Get(shortURL string) (string, error) {
	if m.getErr != nil {
		return "", m.getErr
	}
	longURL, ok := m.data[shortURL]
	if !ok {
		return "", errors.New("not found")
	}
	return longURL, nil
}

type badReader struct{}

func (badReader) Read(p []byte) (int, error) { return 0, errors.New("read error") }
func (badReader) Close() error               { return nil }

func TestCommonHandler_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		body           io.Reader
		repo           *mockRepo
		wantStatus     int
		wantBodyPrefix string
		wantBody       string
		wantLocation   string
	}{
		{
			name:           "POST success",
			method:         http.MethodPost,
			url:            "/",
			body:           strings.NewReader("https://example.com"),
			repo:           &mockRepo{},
			wantStatus:     http.StatusCreated,
			wantBodyPrefix: "http://localhost:8080/",
		},
		{
			name:       "POST read body error",
			method:     http.MethodPost,
			url:        "/",
			body:       badReader{},
			repo:       &mockRepo{},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "POST save error",
			method:     http.MethodPost,
			url:        "/",
			body:       strings.NewReader("https://example.com"),
			repo:       &mockRepo{saveErr: errors.New("save error")},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:         "GET success",
			method:       http.MethodGet,
			url:          "/abc123",
			body:         nil,
			repo:         &mockRepo{data: map[string]string{"abc123": "https://example.com"}},
			wantStatus:   http.StatusTemporaryRedirect,
			wantBody:     "https://example.com",
			wantLocation: "https://example.com",
		},
		{
			name:       "GET not found",
			method:     http.MethodGet,
			url:        "/notfound",
			body:       nil,
			repo:       &mockRepo{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "Method not supported",
			method:     http.MethodPut,
			url:        "/",
			body:       nil,
			repo:       &mockRepo{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "GET repo error",
			method:     http.MethodGet,
			url:        "/abc123",
			body:       nil,
			repo:       &mockRepo{getErr: errors.New("repo error")},
			wantStatus: http.StatusNotFound,
		},
		{
			name:           "POST empty body",
			method:         http.MethodPost,
			url:            "/",
			body:           strings.NewReader(""),
			repo:           &mockRepo{},
			wantStatus:     http.StatusCreated,
			wantBodyPrefix: "http://localhost:8080/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := CommonHandler(tt.repo)
			req := httptest.NewRequest(tt.method, tt.url, tt.body)
			w := httptest.NewRecorder()

			handler(w, req)

			resp := w.Result()
			body, _ := io.ReadAll(resp.Body)
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, resp.StatusCode)
			}
			if tt.wantBodyPrefix != "" && !strings.HasPrefix(string(body), tt.wantBodyPrefix) {
				t.Errorf("expected body prefix %q, got %q", tt.wantBodyPrefix, string(body))
			}
			if tt.wantBody != "" && string(body) != tt.wantBody {
				t.Errorf("expected body %q, got %q", tt.wantBody, string(body))
			}
			if tt.wantLocation != "" && resp.Header.Get("Location") != tt.wantLocation {
				t.Errorf("expected Location header %q, got %q", tt.wantLocation, resp.Header.Get("Location"))
			}
		})
	}
}
