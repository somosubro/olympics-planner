APP_NAME=olympics-planner

.PHONY: run test test-race lint fmt tidy build

run:
	go run ./cmd/api

test:
	go test ./...

test-race:
	go test -race ./...

lint:
	golangci-lint run

fmt:
	gofmt -w .
	goimports -w .

tidy:
	go mod tidy

build:
	mkdir -p bin
	go build -o bin/$(APP_NAME) ./cmd/api
