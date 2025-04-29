.PHONY: all, prepare, run, lint-fix, lint, pack, tests, local_run, clear, new_migration, swagger, docker_run

SHELL := /bin/bash
.SHELLFLAGS := -e -c

go_version = 1.23
main_path = ./cmd/server/main.go
app_name = ./server
swagger_path = ./docs/swagger.json
dc = docker compose

all: clear lint pack swagger
prepare: lint-fix pack swagger tests

lint:
	docker run -t --rm -v $(PWD):/app -w /app golangci/golangci-lint:v2.1.5 golangci-lint run --timeout 5m

lint-fix:
	docker run -t --rm -v $(PWD):/app -w /app golangci/golangci-lint:v2.1.5 golangci-lint run --fix --timeout 5m

pack:
	go build -o $(app_name) $(main_path)

tests:
	go test -v ./...

local_run:
	$(dc) up database rabbitmq -d
	$(app_name)

clear:
	rm -f $(app_name)
	rm -f $(swagger_path)

new_migration:
	docker run --rm -v $(PWD)/migrations:/migrations migrate/migrate create -ext sql -dir /migrations -seq ${filename}

swagger:
	docker run --rm --platform linux/amd64 -v $(PWD):/app -w /app parvez3019/go-swagger3:latest --module-path . --main-file-path $(main_path) --output $(swagger_path) --schema-without-pkg

docker_run:
	$(dc) build
	$(dc) up -d
