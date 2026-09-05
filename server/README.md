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

The API imports `data/exercises.json` into PostgreSQL on first startup. Dataset
GIFs are served by the client from the same checkout under
`/videos/<dataset-gif-file>`.

## Database reference data

PostgreSQL is the runtime source for the exercise catalog and program options.
On the first startup, `data/exercises.json` is imported in a transaction. The
original dataset is unchanged. `reference_imports` records completion; later
startups do not read the JSON or overwrite database edits. GIF files are still
served from the dataset checkout and must remain available for demonstrations.

- `exercise_catalog`: exercise metadata, primary/secondary muscles, equipment,
  multilingual instructions, GIF path and imported media availability.
- `exercise_aliases`: program names and references to catalog exercises. Initial
  names and resolution hints live in migration `002_catalog.sql`; there is no
  `aliases.json`. References are resolved once after the dataset import.
- `program_options`: versioned selection options. Both the options endpoint and
  calculation validation read these values from PostgreSQL on each request.
  Formula math and the eight-week structure remain versioned Go business logic.
  Initial values live only in `003_program_options.sql`. Unit tests supply
  minimal options explicitly; integration tests verify that the SQL seed supports
  the default program and produces the expected baseline plan.

Migrations run in filename order, record their versions in `schema_migrations`,
and acquire a transaction advisory lock to serialize concurrent API startups.
The original idempotent migration is compatible with existing installations.
PostgreSQL must include ICU support (as in the project's PostgreSQL 16 image);
`sport_unicode` makes Russian case-insensitive search independent of OS locale.

The catalog API accepts `query`, `muscle` (primary or secondary), `equipment`,
`bodyPart`, `hasImage`, `limit` (up to 100) and `offset`. Responses include facet
values from the whole catalog, not only the current page. SQL parameters are
bound and `%`/`_` in the search text are treated literally.

Run unit tests with `go test ./...`. To include database integration tests:

```bash
TEST_DATABASE_URL='postgres://sport:sport@localhost:35432/sport?sslmode=disable' go test ./...
```

The integration test uses a unique temporary schema and removes it afterward.
It covers migration replay, dataset import, pagination, filters, Russian search,
unknown IDs, literal search input, preservation of database edits, and reading
updated program options during calculation.
