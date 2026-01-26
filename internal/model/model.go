// Package model содержит структуры данных, используемые в HTTP API и сервисах
// приложения сокращателя ссылок.
package model

// Request представляет собой структуру запроса, содержащую URL-адрес для
// сокращения.
type Request struct {
	URL string `json:"url"`
}

// RequestArrayElement представляет элемент входного массива для batch API.
// Поле CorrelationID используется клиентом для сопоставления ответов.
type RequestArrayElement struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// ResponseArrayElement представляет элемент ответа для batch API.
// Содержит идентификатор корреляции и сгенерированную короткую ссылку.
type ResponseArrayElement struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// ResponsePairElement используется при возврате пар короткая->оригинальная
// для конкретного пользователя.
type ResponsePairElement struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// Response представляет общий JSON-ответ с результатом операции.
type Response struct {
	Result string `json:"result"`
}

type StatsResponse struct {
	URLs  int `json:"urls"`  // количество сокращённых URL в сервисе
	Users int `json:"users"` // количество пользователей в сервисе
}
