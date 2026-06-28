SERVER_DIR := server
SPORT_MAIN_DOMAIN ?= example.com
EXERCISE_MEDIA_BASE_URL ?= https://media.$(SPORT_MAIN_DOMAIN)
EXERCISE_MEDIA_MANIFEST ?= exercise-media.json
S3_PREFIX ?= exercises
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
		'  make sync-media        Clone dataset, upload GIFs to S3, write media manifest' \
		'  make docker-build      Build API Docker image' \
		'  make docker-up         Start API with Docker Compose' \
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
	cd $(SERVER_DIR) && SPORT_MAIN_DOMAIN=$(SPORT_MAIN_DOMAIN) EXERCISE_MEDIA_BASE_URL=$(EXERCISE_MEDIA_BASE_URL) S3_PREFIX=$(S3_PREFIX) go run ./cmd/sync-exercise-media --out $(EXERCISE_MEDIA_MANIFEST)

.PHONY: docker-build
docker-build:
	docker build -t $(DOCKER_IMAGE) -f $(SERVER_DIR)/Dockerfile $(SERVER_DIR)

.PHONY: docker-up
docker-up:
	SPORT_MAIN_DOMAIN=$(SPORT_MAIN_DOMAIN) EXERCISE_MEDIA_BASE_URL=$(EXERCISE_MEDIA_BASE_URL) docker compose up --build api

.PHONY: docker-down
docker-down:
	docker compose down

.PHONY: docker-logs
docker-logs:
	docker compose logs -f api
