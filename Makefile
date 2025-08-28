test:
	go test -coverprofile cover.out ./...
	go tool cover -func cover.out

run:
	go run ./cmd/shortener

build:
	go build -o cmd/shortener/shortener ./cmd/shortener