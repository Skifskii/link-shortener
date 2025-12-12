// Package audit предоставляет простой механизм уведомления наблюдателей о
// событиях аудита (создание ссылок и переход по ним).
package audit

import "time"

const (
	ShortenAction = "shorten"
	FollowAction  = "follow"
)

// AuditService управляет списком наблюдателей и рассылает им события.
type AuditService struct {
	observers []observer
}

type observer interface {
	Update(*Event)
}

// New создаёт новый сервис аудита.
func New() *AuditService {
	return &AuditService{
		observers: make([]observer, 0),
	}
}

// NotifyAll отправляет событие всем зарегистрированным наблюдателям асинхронно.
func (a *AuditService) NotifyAll(e *Event) {
	for _, observer := range a.observers {
		go observer.Update(e)
	}
}

// Register добавляет нового наблюдателя в сервис аудита.
func (a *AuditService) Register(o observer) {
	if o != nil {
		a.observers = append(a.observers, o)
	}
}

// Event описывает событие аудита.
// Поле Timestamp содержит unix-время события, Action указывает тип действия.
type Event struct {
	Timestamp   int64  `json:"ts"`             // unix timestamp события
	Action      string `json:"action"`         // действие: shorten (создание) или follow (прохождение по ссылке)
	UserID      int    `json:"user_id,string"` // идентификатор пользователя, если есть
	OriginalURL string `json:"url"`            // оригинальный (не сокращенный) URL
}

// NewEvent создаёт событие аудита с текущим временем.
func NewEvent(userID int, action, originalURL string) *Event {
	return &Event{
		Timestamp:   time.Now().Unix(),
		Action:      action,
		UserID:      userID,
		OriginalURL: originalURL,
	}
}
