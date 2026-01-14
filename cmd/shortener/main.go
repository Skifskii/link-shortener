package main

import (
	"fmt"

	"github.com/Skifskii/link-shortener/internal/app"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	fmt.Println("Build version:", stringIfNotEmpty(buildVersion))
	fmt.Println("Build date:", stringIfNotEmpty(buildDate))
	fmt.Println("Build commit:", stringIfNotEmpty(buildCommit))

	if err := app.Run(); err != nil {
		panic(err)
	}
}

func stringIfNotEmpty(s string) string {
	if s != "" {
		return s
	}
	return "N/A"
}
