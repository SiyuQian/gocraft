.PHONY: build test run tidy

build:
	go build -o bin/gocraft ./cmd/gocraft

test:
	go test ./...

run:
	go run ./cmd/gocraft

tidy:
	go mod tidy
