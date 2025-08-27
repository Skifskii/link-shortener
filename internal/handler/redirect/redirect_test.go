package redirect

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Skifskii/link-shortener/internal/repository/inmemory"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name         string
		ug           URLGetter
		id           string
		wantStatus   int
		wantLocation string
		wantContains string
	}{
		{
			name: "success",
			ug: &inmemory.InMemoryRepo{
				Store: map[string]string{
					"abc123": "https://example.com",
					"qwe321": "https://www.google.com/",
				},
			},
			id:           "abc123",
			wantStatus:   http.StatusTemporaryRedirect,
			wantLocation: "https://example.com",
			wantContains: "",
		},
		{
			name:         "get error",
			ug:           &inmemory.InMemoryRepo{},
			id:           "abc123",
			wantStatus:   http.StatusNotFound,
			wantLocation: "",
			wantContains: "Ссылка не найдена",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(tt.ug)

			req := httptest.NewRequest(http.MethodGet, "/"+tt.id, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code)
			require.Equal(t, tt.wantLocation, rec.Header().Get("Location"))
		})
	}
}
