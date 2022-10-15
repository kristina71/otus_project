BIN := "./bin/rotation"
DOCKER_IMG="rotation:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd

run: build
	$(BIN) -config ./configs/config.yaml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN) version

test:
	go test -race ./internal/... -count 100

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.41.1


lint: install-lint-deps
    deps: [ install-lint-deps ]
	golangci-lint run ./...

.PHONY: build run build-img run-img version test lint

.PHONY: generate
generate:
	go generate ./...

createdb:
	PGHOST=localhost PGUSER=postgres PGPORT=5432 createdb rotation

migrateUp:
	goose -dir migrations postgres "user=postgres password=password dbname=rotation sslmode=disable" up

migrateDowm:
	goose -dir migrations postgres "user=postgres password=password dbname=rotation sslmode=disable" down

start-evans:
	evans --proto api/*.proto repl