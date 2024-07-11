BIN_AVITOTECH := "./bin/avitotech"

DOCKER_IMG1="avitotech:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

Red='\033[0;31m'
Green='\033[0;32m'
Color_Off='\033[0m'

help:
	@echo ${Red}"Please select a subcommand"${Color_Off}
	@echo ${Green}"make run-postgres"${Color_Off}" to run postgres"
	@echo ${Green}"make create-db"${Color_Off}" to create db"
	@echo ${Green}"make migrate-up"${Color_Off}" to migrate up"
	@echo
	@echo ${Green}"make build"${Color_Off}" to build application"
	@echo ${Green}"make run"${Color_Off}" to run avitotech"
	@echo
	@echo ${Red}"Or use docker-compose:"
	@echo ${Green}"make up"${Color_Off}" to run docker-compose"
	@echo ${Green}"make down"${Color_Off}" to stop docker-compose"
	@echo ${Green}"make destroy"${Color_Off}" to stop docker-compose and remove volumes"
	@echo
	@echo ${Green}"make test"${Color_Off}" to run unit tests"

build: build-avitotech

build-avitotech:
	go build -v -o $(BIN_AVITOTECH) -ldflags "$(LDFLAGS)" ./cmd/avitotech

run: build-avitotech
	$(BIN_AVITOTECH) -config ./configs/config.toml

build-img-avitotech:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG1) \
		-f build/avitotech/Dockerfile .

build-img: build-img-avitotech

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN_AVITOTECH) version

test:
	go clean -testcache;
	go test -race ./internal/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.59.1

lint: install-lint-deps
	golangci-lint run ./...

run-postgres:
	docker run -d --rm --name pg -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secretkey -e PGDATA=/var/lib/postgresql/data/pgdata -v psqldata:/var/lib/postgresql/data -p 5432:5432 postgres:latest

create-db:
	docker exec -it pg createdb --username=root --owner=root avitotech

drop-db:
	docker exec -it pg dropdb avitotech

migrate-up:
	goose -dir migrations postgres "host=localhost user=root password=secretkey dbname=avitotech sslmode=disable" up

migrate-down:
	goose -dir migrations postgres "host=localhost user=root password=secretkey dbname=avitotech sslmode=disable" down

migrate-status:
	goose -dir migrations postgres "host=localhost user=root password=secretkey dbname=avitotech sslmode=disable" status

migrate-reset:
	goose -dir migrations postgres "host=localhost user=root password=secretkey dbname=avitotech sslmode=disable" reset

up:
	@docker-compose -f ./deployments/docker-compose.yaml up -d

down:
	@docker-compose -f ./deployments/docker-compose.yaml down

destroy:
	@docker-compose -f ./deployments/docker-compose.yaml down -v
