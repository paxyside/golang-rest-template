.PHONY: all, prepare, lint-fix, lint, pack, tests, local_run, clear, new_migration, docker_run

SHELL := /bin/bash
.SHELLFLAGS := -e -c

go_version = 1.23
main_path = ./cmd/server/main.go
app_name = ./server
dc = docker compose

all: clear lint pack
prepare: lint-fix pack tests

lint:
	go tool golangci-lint run

lint-fix:
	go tool golangci-lint run --fix

pack:
	go build -o $(app_name) $(main_path)

tests:
	go test -v ./...

local_run:
	$(dc) up database rabbitmq redis prometheus -d
	$(app_name)

clear:
	rm -f $(app_name)

new_migration:
	docker run --rm -v $(PWD)/migrations:/migrations migrate/migrate create -ext sql -dir /migrations -seq ${filename}

docker_run:
	$(dc) build
	$(dc) up -d
