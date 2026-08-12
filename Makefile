SERVER_DIR := server
CLIENT_PORT ?= 38174
API_PORT ?= 38080
POSTGRES_PORT ?= 35432
DOCKER_IMAGE ?= sport-api
DATABASE_URL ?= postgres://sport:sport@localhost:5432/sport?sslmode=disable

.PHONY: help
help:
	@printf '%s\n' \
		'Targets:' \
		'  make server-generate   Generate Go API from docs/openapi.yaml with ogen' \
		'  make server-run        Run API on :8080' \
		'  make server-test       Run Go tests' \
		'  make server-vet        Run go vet' \
		'  make server-race       Run Go race tests' \
		'  make db-up             Start local Postgres only' \
		'  make db-down           Stop local Postgres' \
		'  make docker-build      Build Docker Compose images' \
		'  make docker-up         Start full app with Docker Compose' \
		'  make docker-down       Stop Docker Compose services' \
		'  make docker-logs       Tail API logs'

.PHONY: server-generate
server-generate:
	cd $(SERVER_DIR) && go run github.com/ogen-go/ogen/cmd/ogen@latest --target internal/api --package api --clean ../docs/openapi.yaml

.PHONY: server-run
server-run:
	cd $(SERVER_DIR) && DATABASE_URL='$(DATABASE_URL)' EXERCISE_DATASET_DIR=../exercises-dataset-main go run ./cmd/api

.PHONY: server-test
server-test:
	cd $(SERVER_DIR) && go test ./...

.PHONY: server-vet
server-vet:
	cd $(SERVER_DIR) && go vet ./...

.PHONY: server-race
server-race:
	cd $(SERVER_DIR) && go test -race ./...

.PHONY: db-up
db-up:
	docker compose up -d postgres

.PHONY: db-down
db-down:
	docker compose stop postgres

.PHONY: docker-build
docker-build:
	docker compose build

.PHONY: docker-up
docker-up:
	CLIENT_PORT=$(CLIENT_PORT) API_PORT=$(API_PORT) POSTGRES_PORT=$(POSTGRES_PORT) docker compose up --build

.PHONY: docker-down
docker-down:
	docker compose down

.PHONY: docker-logs
docker-logs:
	docker compose logs -f api
