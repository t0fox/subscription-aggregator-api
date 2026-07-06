.PHONY: run build test lint up down

run:
	go run ./cmd/server

build:
	go build ./cmd/server

test:
	go test -race -cover ./...

lint:
	gofmt -l .
	go vet ./...
	golangci-lint run

up:
	docker compose up --build

down:
	docker compose down
