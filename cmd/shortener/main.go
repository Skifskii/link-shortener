package main

import (
	"net/http"

	"github.com/Skifskii/link-shortener/internal/handler/redirect"
	"github.com/Skifskii/link-shortener/internal/handler/save"
	"github.com/Skifskii/link-shortener/internal/repository"
	"github.com/Skifskii/link-shortener/internal/repository/inmemory"
	"github.com/Skifskii/link-shortener/internal/service/shortener"
	"github.com/go-chi/chi/v5"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	// репозиторий
	var repo repository.Repository = inmemory.New()

	// генератор коротких ссылок
	s := shortener.New(6)

	// роутер
	r := chi.NewRouter()

	r.Get("/{id}", redirect.New(repo))
	r.Post("/", save.New(repo, s))

	return http.ListenAndServe(`:8080`, r)
}
