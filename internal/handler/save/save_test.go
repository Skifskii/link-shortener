package save

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

type mockShortener struct {
	shortURL string
	err      error
}

func (ms *mockShortener) Shorten(_ string) (shortURL string, err error) {
	return ms.shortURL, ms.err
}

// func TestNew(t *testing.T) {
// 	tests := []struct {
// 		name string // description of this test case
// 		// Named input parameters for target function.
// 		s            Shortener
// 		body         string
// 		customBody   io.ReadCloser
// 		wantStatus   int
// 		wantContains string
// 	}{
// 		{
// 			name:         "success",
// 			body:         "https://example.com",
// 			s:            &mockShortener{"http://localhost:8080/abc123", nil},
// 			wantStatus:   http.StatusCreated,
// 			wantContains: "http://localhost:8080/abc123",
// 		},
// 		{
// 			name:         "read body error",
// 			customBody:   &errBody{err: errors.New("read error")},
// 			wantStatus:   http.StatusInternalServerError,
// 			wantContains: "Не удалось прочитать тело запроса",
// 		},
// 		{
// 			name:         "generate short link error",
// 			body:         "https://example.com",
// 			s:            &mockShortener{"", errors.New("error")},
// 			wantStatus:   http.StatusInternalServerError,
// 			wantContains: "Ошибка при генерации короткой ссылки",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			h := New(tt.s)

// 			req := httptest.NewRequest(http.MethodPost, "/", nil)

// 			if tt.customBody != nil {
// 				req.Body = tt.customBody
// 			} else {
// 				req.Body = io.NopCloser(strings.NewReader(tt.body))
// 			}

// 			rec := httptest.NewRecorder()

// 			h.ServeHTTP(rec, req)

// 			require.Equal(t, tt.wantStatus, rec.Code)
// 			require.Contains(t, rec.Body.String(), tt.wantContains)
// 		})
// 	}
// }
