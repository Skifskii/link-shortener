package model

type Request struct {
	URL string `json:"url"`
}

type RequestArrayElement struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ResponseArrayElement struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type ResponsePairElement struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Response struct {
	Result string `json:"result"`
}
