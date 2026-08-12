# Server

Go backend generated from `../docs/openapi.yaml` with `ogen`.

## Generate API

```bash
go run github.com/ogen-go/ogen/cmd/ogen@latest --target internal/api --package api --clean ../docs/openapi.yaml
```

## Generate Wire Injector

```bash
go generate ./cmd/api
```

Or directly:

```bash
wire gen ./cmd/api
```

## Run API

```bash
DATABASE_URL='postgres://sport:sport@localhost:5432/sport?sslmode=disable' go run ./cmd/api
```

Useful environment variables:

- `SPORT_API_ADDR`: listen address, default `:8080`.
- `DATABASE_URL`: PostgreSQL connection string.
- `DATABASE_RUN_MIGRATIONS`: apply embedded migrations on startup, default `true`.
- `EXERCISE_DATASET_DIR`: path to `exercises-dataset-main`, default `../exercises-dataset-main`.
- `LOG_LEVEL`, `LOG_PRETTY`, `HTTP_TIMEOUT`, `HTTP_REQUEST_ID_HEADER`: loaded through `configx`.

Configuration uses `github.com/dkoshenkov/packages-go/configx`.
Logging uses `github.com/dkoshenkov/packages-go/logx`.
HTTP request ID, timeout, recovery, and logging middleware use `github.com/dkoshenkov/packages-go/middlewarex/httpx`.

The API reads `data/exercises.json` directly on startup. Dataset GIFs are
served from the same checkout under `/videos/<dataset-gif-file>`; no import,
clone, manifest, or media synchronization step is required.
