package redirect

import (
	"errors"
	"net/http"

	"github.com/Skifskii/link-shortener/internal/middleware/authmw"
	"github.com/Skifskii/link-shortener/internal/repository"
	"github.com/Skifskii/link-shortener/internal/service/audit"
	"github.com/go-chi/chi/v5"
)

type ShortRedirecter interface {
	Redirect(shortURL string) (longURL string, err error)
}

type AuditEventNotifier interface {
	NotifyAll(*audit.Event)
}

func New(sr ShortRedirecter, auditEventNotifier AuditEventNotifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortURL := chi.URLParam(r, "id")
		if shortURL == "" {
			http.Error(w, "id не указан", http.StatusBadRequest)
			return
		}

		longURL, err := sr.Redirect(shortURL)
		if err != nil {
			if errors.Is(err, repository.ErrLinkDeleted) {
				w.WriteHeader(http.StatusGone)
				return
			}
			http.Error(w, "Ссылка не найдена", http.StatusNotFound)
			return
		}

		w.Header().Set("Location", longURL)
		w.WriteHeader(http.StatusTemporaryRedirect)

		// После успешного запроса отправляем уведомление
		userID := r.Context().Value(authmw.UserIDKey).(int)
		auditEventNotifier.NotifyAll(audit.NewEvent(userID, audit.FollowAction, longURL))
	}
}
