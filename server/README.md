# Server

Go backend generated from `../docs/openapi.yaml` with `ogen`.

## Generate API

```bash
go run github.com/ogen-go/ogen/cmd/ogen@latest --target internal/api --package api --clean ../docs/openapi.yaml
```

## Run API

```bash
DATABASE_URL='postgres://sport:sport@localhost:5432/sport?sslmode=disable' go run ./cmd/api
```

Useful environment variables:

- `SPORT_API_ADDR`: listen address, default `:8080`.
- `DATABASE_URL`: PostgreSQL connection string.
- `DATABASE_RUN_MIGRATIONS`: apply embedded migrations on startup, default `true`.
- `SPORT_MAIN_DOMAIN`: main domain used to derive `https://media.<domain>`, default `example.com`.
- `EXERCISE_MEDIA_BASE_URL`: optional override for GIF public base URL.
- `EXERCISE_MEDIA_MANIFEST`: optional path to a generated exercise media JSON manifest.
- `LOG_LEVEL`, `LOG_PRETTY`, `HTTP_TIMEOUT`, `HTTP_REQUEST_ID_HEADER`: loaded through `configx`.

Configuration uses `github.com/dkoshenkov/packages-go/configx`.
Logging uses `github.com/dkoshenkov/packages-go/logx`.
HTTP request ID, timeout, recovery, and logging middleware use `github.com/dkoshenkov/packages-go/middlewarex/httpx`.

## Sync Exercise GIFs to S3

```bash
S3_BUCKET=my-bucket \
SPORT_MAIN_DOMAIN=example.com \
go run ./cmd/sync-exercise-media --out exercise-media.json
```

The sync command clones `hasaneyldrm/exercises-dataset` into a temporary directory,
matches only the program alias layer, uploads GIFs to `s3://$S3_BUCKET/exercises/`,
and writes a manifest consumable by `EXERCISE_MEDIA_MANIFEST`.

By default manifest URLs use `https://media.<SPORT_MAIN_DOMAIN>/exercises/...`.
Set `EXERCISE_MEDIA_BASE_URL` to override that, for example to a CloudFront URL.
