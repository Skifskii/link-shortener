package router

import (
	"fmt"
	"net/http"

	"github.com/Skifskii/link-shortener/internal/handler/batch"
	"github.com/Skifskii/link-shortener/internal/handler/delete"
	"github.com/Skifskii/link-shortener/internal/handler/ping"
	"github.com/Skifskii/link-shortener/internal/handler/redirect"
	"github.com/Skifskii/link-shortener/internal/handler/save"
	"github.com/Skifskii/link-shortener/internal/handler/shorten"
	"github.com/Skifskii/link-shortener/internal/handler/urls"
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
	r.Get("/{id}", redirect.New(shorter, aud))
	r.Post("/", save.New(shorter, aud))
	r.Post("/api/shorten", shorten.New(shorter, aud))
	r.Get("/ping", ping.New(p))
	r.Post("/api/shorten/batch", batch.New(shorter))
	r.Get("/api/user/urls", urls.New(shorter))
	r.Delete("/api/user/urls", delete.New(shorter))

	return &Router{r}
}

func (r *Router) Run(address string) error {
	fmt.Printf("Starting server at %s\n", address)
	return http.ListenAndServe(address, r.chiRouter)
}
