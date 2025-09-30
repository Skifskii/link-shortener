package redirect

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

// func TestNewOld(t *testing.T) {
// 	tests := []struct {
// 		name         string
// 		ug           URLGetter
// 		id           string
// 		wantStatus   int
// 		wantLocation string
// 		wantContains string
// 	}{
// 		{
// 			name: "success",
// 			ug: &inmemory.InMemoryRepo{
// 				Store: map[string]string{
// 					"abc123": "https://example.com",
// 					"qwe321": "https://www.google.com/",
// 				},
// 			},
// 			id:           "abc123",
// 			wantStatus:   http.StatusTemporaryRedirect,
// 			wantLocation: "https://example.com",
// 			wantContains: "",
// 		},
// 		{
// 			name:         "get error",
// 			ug:           &inmemory.InMemoryRepo{},
// 			id:           "abc123",
// 			wantStatus:   http.StatusNotFound,
// 			wantLocation: "",
// 			wantContains: "Ссылка не найдена",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			h := New(tt.ug)

// 			req := httptest.NewRequest(http.MethodGet, "/"+tt.id, nil)

// 			rctx := chi.NewRouteContext()
// 			rctx.URLParams.Add("id", tt.id)
// 			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

// 			rec := httptest.NewRecorder()

// 			h.ServeHTTP(rec, req)

// 			require.Equal(t, tt.wantStatus, rec.Code)
// 			require.Equal(t, tt.wantLocation, rec.Header().Get("Location"))
// 		})
// 	}
// }

type mockShortRedirecter struct {
	longURL string
	err     error
}

func (ms *mockShortRedirecter) Redirect(_ string) (longURL string, err error) {
	return ms.longURL, ms.err
}

func TestNew(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		sr           ShortRedirecter
		id           string
		wantStatus   int
		wantLocation string
		wantContains string
	}{
		{
			name:         "success",
			sr:           &mockShortRedirecter{"https://example.com", nil},
			id:           "abc123",
			wantStatus:   http.StatusTemporaryRedirect,
			wantLocation: "https://example.com",
			wantContains: "",
		},
		{
			name:         "get error",
			sr:           &mockShortRedirecter{"", errors.New("error")},
			id:           "abc123",
			wantStatus:   http.StatusNotFound,
			wantLocation: "",
			wantContains: "Ссылка не найдена",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(tt.sr)

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
