package router

import (
	"fmt"
	"net/http"

	"github.com/Skifskii/link-shortener/internal/handler/batch"
	"github.com/Skifskii/link-shortener/internal/handler/ping"
	"github.com/Skifskii/link-shortener/internal/handler/redirect"
	"github.com/Skifskii/link-shortener/internal/handler/save"
	"github.com/Skifskii/link-shortener/internal/handler/shorten"
	"github.com/Skifskii/link-shortener/internal/logger"
	"github.com/Skifskii/link-shortener/internal/middleware"
	"github.com/Skifskii/link-shortener/internal/model"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Router struct {
	chiRouter *chi.Mux
}

type Shorter interface {
	Shorten(longURL string) (shortURL string, err error)
	Redirect(shortURL string) (longURL string, err error)
	BatchShorten(reqBatch []model.RequestArrayElement) (respBatch []model.ResponseArrayElement, err error)
}

type pinger interface {
	Ping() error
}

func New(zl *zap.Logger, shorter Shorter, p pinger) *Router {
	r := chi.NewRouter()

	r.Use(logger.RequestLogger(zl))
	r.Use(middleware.GzipMiddleware)

	r.Get("/{id}", redirect.New(shorter))
	r.Post("/", save.New(shorter))
	r.Post("/api/shorten", shorten.New(shorter))
	r.Get("/ping", ping.New(p))
	r.Post("/api/shorten/batch", batch.New(shorter))

	return &Router{r}
}

func (r *Router) Run(address string) error {
	fmt.Printf("Starting server at %s\n", address)
	return http.ListenAndServe(address, r.chiRouter)
}
