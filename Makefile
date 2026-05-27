.PHONY: build install test run tidy

build:
	go build -o bin/gocraft ./cmd/gocraft

install:
	go install ./cmd/gocraft

test:
	go test ./...

run:
	go run ./cmd/gocraft

tidy:
	go mod tidy
