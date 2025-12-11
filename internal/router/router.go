package router

import (
	"fmt"
	"net/http"

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

type Router struct {
	chiRouter *chi.Mux
}

type Shorter interface {
	Shorten(userID int, longURL string) (shortURL string, err error)
	Redirect(shortURL string) (longURL string, err error)
	BatchShorten(reqBatch []model.RequestArrayElement) (respBatch []model.ResponseArrayElement, err error)
	GetUserPairs(userID int) ([]model.ResponsePairElement, error)
	DeleteUserLinks(userID int, shortURLs []string) error
}

type pinger interface {
	Ping() error
}

type Auther interface {
	CreateUser(username string) (jwt string, err error)
	GetUserID(tokenString string) (int, error)
}

type auditEventNotifier interface {
	NotifyAll(*audit.Event)
}

func New(zl *zap.Logger, shorter Shorter, p pinger, auth Auther, aud auditEventNotifier) *Router {
	r := chi.NewRouter()

	// middlewares
	r.Use(logger.RequestLogger(zl))
	r.Use(authmw.AuthMiddleware(auth))
	r.Use(gzipmw.GzipMiddleware)

	// handlers
	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", redirect.New(shorter, aud))
		r.Post("/", save.New(shorter, aud))
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
	})

	r.Get("/ping", ping.New(p))

	return &Router{r}
}

func (r *Router) Run(address string) error {
	fmt.Printf("Starting server at %s\n", address)
	return http.ListenAndServe(address, r.chiRouter)
}
