package main

import (
	"net/http"

	"github.com/Skifskii/link-shortener/internal/handler"
	"github.com/Skifskii/link-shortener/internal/repository"
	"github.com/Skifskii/link-shortener/internal/repository/inmemory"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	var repo repository.Repository = inmemory.New()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.CommonHandler(repo))

	return http.ListenAndServe(`:8080`, mux)
}
