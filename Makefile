SERVER_DIR := server
SPORT_MAIN_DOMAIN ?= example.com
EXERCISE_DATASET_DIR ?= exercises-dataset
EXERCISE_MEDIA_SOURCE_DIR ?= exercises-gifs
EXERCISE_MEDIA_BASE_URL ?= https://media.$(SPORT_MAIN_DOMAIN)
EXERCISE_MEDIA_MANIFEST ?= exercise-media.json
S3_PREFIX ?= exercises
EXERCISE_MEDIA_STORAGE_MODE ?= local
EXERCISE_MEDIA_LOCAL_DIR ?= ../client/public/exercises
CLIENT_PORT ?= 38174
API_PORT ?= 38080
POSTGRES_PORT ?= 35432
DOCKER_EXERCISE_MEDIA_BASE_URL ?=
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
		'  make sync-media        Clone dataset, seed DB, sync GIF media, write media manifest' \
		'  make docker-sync-media Clone dataset, seed DB, sync GIF media through Docker Compose' \
		'  make docker-build      Build Docker Compose images' \
		'  make docker-up         Start full app with Docker Compose' \
		'  make docker-down       Stop Docker Compose services' \
		'  make docker-logs       Tail API logs'

.PHONY: server-generate
server-generate:
	cd $(SERVER_DIR) && go run github.com/ogen-go/ogen/cmd/ogen@latest --target internal/api --package api --clean ../docs/openapi.yaml

.PHONY: server-run
server-run:
	cd $(SERVER_DIR) && DATABASE_URL='$(DATABASE_URL)' SPORT_MAIN_DOMAIN=$(SPORT_MAIN_DOMAIN) EXERCISE_MEDIA_BASE_URL=$(EXERCISE_MEDIA_BASE_URL) go run ./cmd/api

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

.PHONY: sync-media
sync-media:
	cd $(SERVER_DIR) && DATABASE_URL='$(DATABASE_URL)' SPORT_MAIN_DOMAIN=$(SPORT_MAIN_DOMAIN) EXERCISE_DATASET_DIR=../$(EXERCISE_DATASET_DIR) EXERCISE_MEDIA_BASE_URL=$(EXERCISE_MEDIA_BASE_URL) EXERCISE_MEDIA_SOURCE_DIR=../$(EXERCISE_MEDIA_SOURCE_DIR) EXERCISE_MEDIA_SOURCE_BASE_URL=$(EXERCISE_MEDIA_SOURCE_BASE_URL) EXERCISE_MEDIA_STORAGE_MODE=$(EXERCISE_MEDIA_STORAGE_MODE) EXERCISE_MEDIA_LOCAL_DIR=$(EXERCISE_MEDIA_LOCAL_DIR) EXERCISE_MEDIA_MANIFEST=$(EXERCISE_MEDIA_MANIFEST) S3_PREFIX=$(S3_PREFIX) go run ./cmd/sync-exercise-media

.PHONY: docker-sync-media
docker-sync-media:
	SPORT_MAIN_DOMAIN=$(SPORT_MAIN_DOMAIN) CLIENT_PORT=$(CLIENT_PORT) POSTGRES_PORT=$(POSTGRES_PORT) docker compose --profile tools build media-sync
	SPORT_MAIN_DOMAIN=$(SPORT_MAIN_DOMAIN) CLIENT_PORT=$(CLIENT_PORT) POSTGRES_PORT=$(POSTGRES_PORT) EXERCISE_DATASET_DIR=/dataset EXERCISE_MEDIA_BASE_URL=$(DOCKER_EXERCISE_MEDIA_BASE_URL) EXERCISE_MEDIA_SOURCE_DIR=/gifs EXERCISE_MEDIA_SOURCE_BASE_URL=$(EXERCISE_MEDIA_SOURCE_BASE_URL) EXERCISE_MEDIA_STORAGE_MODE=$(EXERCISE_MEDIA_STORAGE_MODE) S3_PREFIX=$(S3_PREFIX) docker compose --profile tools run --rm media-sync --exercise-media-manifest /media/exercises/exercise-media.json

.PHONY: docker-build
docker-build:
	docker compose build

.PHONY: docker-up
docker-up:
	SPORT_MAIN_DOMAIN=$(SPORT_MAIN_DOMAIN) CLIENT_PORT=$(CLIENT_PORT) API_PORT=$(API_PORT) POSTGRES_PORT=$(POSTGRES_PORT) docker compose up --build

.PHONY: docker-down
docker-down:
	docker compose down

.PHONY: docker-logs
docker-logs:
	docker compose logs -f api
