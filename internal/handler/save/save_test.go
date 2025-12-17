package save

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Skifskii/link-shortener/internal/handler/save/mocks"
	"github.com/Skifskii/link-shortener/internal/middleware/authmw"
	"github.com/Skifskii/link-shortener/internal/repository/inmemory"
	"github.com/Skifskii/link-shortener/internal/service/auth"
	"github.com/Skifskii/link-shortener/internal/service/shortener"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	one := 1
	tests := []struct {
		name            string
		userID          *int
		body            string
		shortenShortURL string
		shortenErr      error
		wantStatus      int
		wantContains    string
	}{
		{
			name:            "success",
			userID:          &one,
			wantStatus:      http.StatusCreated,
			shortenShortURL: "abc",
		},
		{
			name:       "failed to get userID",
			userID:     nil,
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortenerMock := mocks.NewShortener(t)
			if tt.userID != nil {
				shortenerMock.On("Shorten", mock.Anything, tt.body).
					Return(tt.shortenShortURL, tt.shortenErr)
			}

			auditEventNotifierMock := mocks.NewAuditEventNotifier(t)
			var notifyDone chan struct{}
			if tt.wantStatus == http.StatusCreated {
				notifyDone = make(chan struct{})
				auditEventNotifierMock.
					On("NotifyAll", mock.Anything).
					Run(func(args mock.Arguments) {
						// сигнализируем, что вызов произошёл
						close(notifyDone)
					}).
					Return()
			}

			handler := New(shortenerMock, auditEventNotifierMock)

			req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			if tt.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), authmw.UserIDKey, *tt.userID))
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code)
			require.Contains(t, rec.Body.String(), tt.wantContains)

			if notifyDone != nil {
				select {
				case <-notifyDone:
				case <-time.After(1 * time.Second):
					t.Fatal("timeout waiting for NotifyAll to be called")
				}
			}
		})
	}
}

func BenchmarkNew(b *testing.B) {
	repo := inmemory.New()

	shorter := shortener.New("localhost:8080", 6, repo)

	auditService := &mocks.AuditEventNotifier{}
	auditService.On("NotifyAll", mock.Anything).Return()

	authServiece := auth.New(repo, "secret")

	router := chi.NewRouter()
	router.Use(authmw.AuthMiddleware(authServiece))
	router.Post("/", New(shorter, auditService))

	body := "https://example.com/test"

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
	}
}
