package main

import "github.com/Skifskii/link-shortener/internal/app"

func main() {
	if err := app.Run(); err != nil {
		panic(err)
	}
}
