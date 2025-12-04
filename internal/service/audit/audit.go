package audit

import "time"

const (
	ShortenAction = "shorten"
	FollowAction  = "follow"
)

type AuditService struct {
	observers []observer
}

type observer interface {
	Update(*Event)
}

func New() *AuditService {
	return &AuditService{
		observers: make([]observer, 0),
	}
}

func (a *AuditService) NotifyAll(e *Event) {
	for _, observer := range a.observers {
		observer.Update(e)
	}
}

func (a *AuditService) Register(o observer) {
	if o != nil {
		a.observers = append(a.observers, o)
	}
}

type Event struct {
	Timestamp   int64  `json:"ts"`             // unix timestamp события
	Action      string `json:"action"`         // действие: shorten (создание) или follow (прохождение по ссылке)
	UserID      int    `json:"user_id,string"` // идентификатор пользователя, если есть
	OriginalURL string `json:"url"`            // оригинальный (не сокращенный) URL
}

func NewEvent(userID int, action, originalURL string) *Event {
	return &Event{
		Timestamp:   time.Now().Unix(),
		Action:      action,
		UserID:      userID,
		OriginalURL: originalURL,
	}
}
