package batch

import (
	"encoding/json"
	"net/http"

	"github.com/Skifskii/link-shortener/internal/model"
)

type BatchShortener interface {
	BatchShorten(reqBatch []model.RequestArrayElement) (respBatch []model.ResponseArrayElement, err error)
}

func New(bs BatchShortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		dec := json.NewDecoder(r.Body)

		// проверяем, что json начинается со скобки '['
		t, err := dec.Token()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if delim, ok := t.(json.Delim); !ok || delim != '[' {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// готовим выходной набор
		respArray := make([]model.ResponseArrayElement, 0)

		// парсим входной набор
		const maxBatchSize = 1_000
		reqBatch := make([]model.RequestArrayElement, 0, maxBatchSize)

		for dec.More() {
			var el model.RequestArrayElement
			if err := dec.Decode(&el); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			reqBatch = append(reqBatch, el)

			if len(reqBatch) >= maxBatchSize {
				respBatch, err := bs.BatchShorten(reqBatch)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				respArray = append(respArray, respBatch...)
				reqBatch = reqBatch[:0]
			}
		}

		if len(reqBatch) > 0 {
			respBatch, err := bs.BatchShorten(reqBatch)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			respArray = append(respArray, respBatch...)
		}

		// сериализуем ответ сервера
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		enc := json.NewEncoder(w)
		enc.Encode(respArray)
	}
}
