BINARY := crds

.PHONY: all build install test lint tidy legacy run

all: build

build:
	go build -o $(BINARY) ./cmd/crds/

install:
	go install ./cmd/crds/

test:
	go test ./...

lint:
	golangci-lint run

tidy:
	go mod tidy

legacy:
	go build -o crds-legacy ./cmd/legacy-quiz/

run:
	go run ./cmd/crds/
