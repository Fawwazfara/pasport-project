BINARY := passport
PKG := ./cmd/server

.PHONY: all vet test build run docker clean

all: vet test build

vet:
	go vet ./...

test:
	go test -race ./...

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/$(BINARY) $(PKG)

run:
	PORT=8080 STORAGE_DIR=./storage go run $(PKG)

docker:
	docker compose up -d --build

clean:
	rm -rf bin storage
