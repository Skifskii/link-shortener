// Package router конфигурирует маршруты HTTP и связывает их с обработчиками.
package router

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Skifskii/link-shortener/internal/handler/api/inter/stats"
	"github.com/Skifskii/link-shortener/internal/handler/api/shorten"
	"github.com/Skifskii/link-shortener/internal/handler/api/shorten/batch"
	"github.com/Skifskii/link-shortener/internal/handler/api/user/urls"
	"github.com/Skifskii/link-shortener/internal/handler/ping"
	"github.com/Skifskii/link-shortener/internal/handler/redirect"
	"github.com/Skifskii/link-shortener/internal/handler/save"

	"github.com/Skifskii/link-shortener/internal/logger"
	"github.com/Skifskii/link-shortener/internal/middleware/authmw"
	"github.com/Skifskii/link-shortener/internal/middleware/gzipmw"
	"github.com/Skifskii/link-shortener/internal/model"
	"github.com/Skifskii/link-shortener/internal/service/audit"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Router обёртка над chi-маршрутизатором.
type Router struct {
	chiRouter *chi.Mux
}

// Shorter интерфейс для сокращения ссылок, объединяющий методы, которые
// используются различными хендлерами приложения.
type Shorter interface {
	Shorten(userID int, longURL string) (shortURL string, err error)
	Redirect(shortURL string) (longURL string, err error)
	BatchShorten(userID int, reqBatch []model.RequestArrayElement) (respBatch []model.ResponseArrayElement, err error)
	GetUserPairs(userID int) ([]model.ResponsePairElement, error)
	DeleteUserLinks(userID int, shortURLs []string) error
}

// pinger интерфейс для проверки доступности сервиса (используется в /ping handler).
type pinger interface {
	Ping() error
}

// Auther интерфейс для работы с пользователями и JWT-токенами.
type Auther interface {
	CreateUser(username string) (jwt string, err error)
	GetUserID(tokenString string) (int, error)
}

// auditEventNotifier интерфейс для уведомления о событиях аудита.
type auditEventNotifier interface {
	NotifyAll(*audit.Event)
}

// statsGetter - интерфейс для получения статистики.
type statsGetter interface {
	GetIfAllowed(ip net.IP) (model.StatsResponse, error)
}

// New создаёт новый Router, регистрирует middleware и обработчики.
func New(zl *zap.Logger, shorter Shorter, p pinger, auth Auther, aud auditEventNotifier, stat statsGetter) *Router {
	r := chi.NewRouter()

	// middlewares
	r.Use(logger.RequestLogger(zl))
	r.Use(authmw.AuthMiddleware(auth))
	r.Use(gzipmw.GzipMiddleware)

	// handlers
	r.Route("/", func(r chi.Router) {
		r.Post("/", save.New(shorter, aud))
		r.Get("/{id}", redirect.New(shorter, aud))
	})
	r.Route("/api", func(r chi.Router) {
		r.Route("/shorten", func(r chi.Router) {
			r.Post("/", shorten.New(shorter, aud))
			r.Post("/batch", batch.New(shorter))
		})
		r.Route("/user", func(r chi.Router) {
			r.Get("/urls", urls.New(shorter))
			r.Delete("/urls", urls.NewDelete(shorter))
		})
		r.Route("/internal", func(r chi.Router) {
			r.Get("/stats", stats.New(stat))
		})
	})
	r.Get("/ping", ping.New(p))

	return &Router{r}
}

// Run запускает HTTP сервер на указанном адресе.
func (r *Router) Run(address, certPath, keyPath string, enableHTTPS bool) error {
	server := &http.Server{
		Addr:    address,
		Handler: r.chiRouter,
	}

	// через этот канал сообщим основному потоку, что соединения закрыты
	connsClosed := make(chan struct{})

	// канал для перенаправления прерываний
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// запускаем горутину обработки пойманных прерываний
	go func() {
		<-sigint
		fmt.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			fmt.Printf("HTTP server Shutdown: %v\n", err)
		}

		close(connsClosed)
	}()

	// Запускаем сервер
	fmt.Printf("Starting HTTP server at %s\n", address)
	var err error
	if enableHTTPS {
		err = r.RunTLS(server, certPath, keyPath)
	} else {
		err = server.ListenAndServe()
	}
	if err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server ListenAndServe: %w", err)
	}

	// Ждём закрытия всех соединений
	<-connsClosed
	fmt.Println("Server Shutdown gracefully")

	return nil
}

// RunTLS - запускает HTTPS сервер на указанном адресе с заданными сертификатом и ключом.
func (*Router) RunTLS(server *http.Server, certPath, keyPath string) error {
	if certPath == "" || keyPath == "" {
		return fmt.Errorf("TLS certificate path and key path must be provided for HTTPS")
	}

	return server.ListenAndServeTLS(certPath, keyPath)
}
