package urls

import (
	"encoding/json"
	"net/http"

	"github.com/Skifskii/link-shortener/internal/middleware/authmw"
	"github.com/Skifskii/link-shortener/internal/model"
)

type Shortener interface {
	GetUserPairs(userID int) ([]model.ResponsePairElement, error)
}

func New(s Shortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		userID, ok := r.Context().Value(authmw.UserIDKey).(int)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		pairs, err := s.GetUserPairs(userID)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if len(pairs) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// сериализуем ответ сервера
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		enc := json.NewEncoder(w)
		enc.Encode(pairs)
	}
}
