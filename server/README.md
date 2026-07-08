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
- `SPORT_MAIN_DOMAIN`: main domain used to derive `https://media.<domain>`, default `example.com`.
- `EXERCISE_MEDIA_BASE_URL`: optional override for GIF public base URL.
- `EXERCISE_MEDIA_MANIFEST`: optional path to a generated exercise media JSON manifest.
- `LOG_LEVEL`, `LOG_PRETTY`, `HTTP_TIMEOUT`, `HTTP_REQUEST_ID_HEADER`: loaded through `configx`.

Configuration uses `github.com/dkoshenkov/packages-go/configx`.
Logging uses `github.com/dkoshenkov/packages-go/logx`.
HTTP request ID, timeout, recovery, and logging middleware use `github.com/dkoshenkov/packages-go/middlewarex/httpx`.

## Sync Exercise Dataset and GIFs

```bash
DATABASE_URL='postgres://sport:sport@localhost:35432/sport?sslmode=disable' \
EXERCISE_DATASET_DIR=../exercises-dataset \
EXERCISE_MEDIA_BASE_URL='http://127.0.0.1:38174' \
EXERCISE_MEDIA_SOURCE_DIR=../exercises-gifs \
EXERCISE_MEDIA_STORAGE_MODE=local \
EXERCISE_MEDIA_LOCAL_DIR=../client/public/exercises \
EXERCISE_MEDIA_MANIFEST=../client/public/exercises/exercise-media.json \
go run ./cmd/sync-exercise-media
```

With Docker Compose:

```bash
make docker-sync-media
```

The sync command clones `hasaneyldrm/exercises-dataset`, imports all dataset
records into Postgres, applies the program alias/RU override layer, downloads
GIFs, and writes media rows into `exercise_media`.

Set `EXERCISE_DATASET_DIR` to read an already cloned local dataset instead of
cloning GitHub on each run. Set `EXERCISE_MEDIA_SOURCE_DIR` to a local GIF
mirror; the sync command looks for `assets/<dataset_id>.gif`,
`<dataset_id>.gif`, `<media_id>.gif`, and old `gif_url` basenames. If no local
file exists, `EXERCISE_MEDIA_SOURCE_BASE_URL` can still point to a licensed
mirror/CDN that serves `${media_id}.gif`. Local mode stores files under
`client/public/exercises` and exposes them as
`${EXERCISE_MEDIA_BASE_URL}/exercises/<dataset_id>.gif`.

For S3 mode, set `EXERCISE_MEDIA_STORAGE_MODE=s3` and `S3_BUCKET=my-bucket`.
By default manifest URLs use `https://media.<SPORT_MAIN_DOMAIN>/exercises/...`;
set `EXERCISE_MEDIA_BASE_URL` to override that, for example to CloudFront.
