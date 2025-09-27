package save

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// мок для генератора коротких ссылок
type mockGenerator struct {
	err   error
	fixed string
}

func (mg *mockGenerator) GenerateShort() (string, error) {
	return mg.fixed, mg.err
}

// мок для сохранения ссылок
type mockSaver struct {
	err        error
	savedShort string
	savedLong  string
}

func (ms *mockSaver) Save(shortURL, longURL string) error {
	ms.savedShort = shortURL
	ms.savedLong = longURL

	return ms.err
}

// мок для симуляции ошибки чтения тела запроса
type errBody struct {
	err error
}

func (e *errBody) Read(p []byte) (int, error) {
	return 0, e.err
}

func (e *errBody) Close() error {
	return nil
}

func TestNew(t *testing.T) {
	baseURL := "http://localhost:8080"

	tests := []struct {
		name         string
		body         string
		customBody   io.ReadCloser
		gen          *mockGenerator
		saver        *mockSaver
		wantStatus   int
		wantContains string
		wantSave     bool
	}{
		{
			name:         "success",
			body:         "https://example.com",
			gen:          &mockGenerator{fixed: "abc123"},
			saver:        &mockSaver{},
			wantStatus:   http.StatusCreated,
			wantContains: "http://localhost:8080/abc123",
			wantSave:     true,
		},
		{
			name:         "read body error",
			customBody:   &errBody{err: errors.New("read error")},
			gen:          &mockGenerator{fixed: "abc123"},
			saver:        &mockSaver{},
			wantStatus:   http.StatusInternalServerError,
			wantContains: "Не удалось прочитать тело запроса",
			wantSave:     false,
		},
		{
			name:         "generate short link error",
			body:         "https://example.com",
			gen:          &mockGenerator{fixed: "abc123", err: errors.New("gen error")},
			saver:        &mockSaver{},
			wantStatus:   http.StatusInternalServerError,
			wantContains: "Ошибка при генерации короткой ссылки",
			wantSave:     false,
		},
		{
			name:         "save error",
			body:         "https://example.com",
			gen:          &mockGenerator{fixed: "abc123"},
			saver:        &mockSaver{err: errors.New("save error")},
			wantStatus:   http.StatusInternalServerError,
			wantContains: "Ошибка при сохранении ссылки",
			wantSave:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(tt.saver, tt.gen, baseURL)

			req := httptest.NewRequest(http.MethodPost, "/", nil)

			if tt.customBody != nil {
				req.Body = tt.customBody
			} else {
				req.Body = io.NopCloser(strings.NewReader(tt.body))
			}

			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code)
			require.Contains(t, rec.Body.String(), tt.wantContains)

			if tt.wantSave {
				require.Equal(t, tt.gen.fixed, tt.saver.savedShort)
				require.Equal(t, tt.body, tt.saver.savedLong)
			}
		})
	}
}
