# Service Spec

This directory is the SDD source of truth for the backend service contract.

## Files

- `openapi.yaml` - OpenAPI 3.1.0 contract intended for `ogen`.
- `data-model.md` - database-oriented domain model.
- `xlsx-analysis.md` - extracted XLSX behavior and formula notes.
- `ui-decisions.md` - product/UI decisions that affect the contract.

## Generation Target

Use `ogen` against `docs/openapi.yaml`.

Example:

```bash
go run github.com/ogen-go/ogen/cmd/ogen --target server/internal/api --package api --clean docs/openapi.yaml
```

## Service Boundary

The service owns:

- local nickname/password authentication;
- athlete profile defaults;
- program options exposed to the client;
- training day calculation from selected 1RM/week/variant/options;
- exercise details for program-relevant exercises;
- stable media URLs for exercise GIF previews;
- bootstrap metadata for the client session;
- persisted user cycle settings;
- persisted training progress checkpoints.

The service does not own:

- SSO login;
- billing or account management;
- SSO profile management.

Users are local to this service:

- `nickname` is unique;
- password is accepted only on register/login;
- only a password hash is stored;
- user state is attached to the authenticated nickname account.

## Cookie Contract

Current bootstrap cookie:

- Name: `init`
- Value: `1`
- Purpose: temporary client bootstrap marker
- JavaScript-readable: yes
- `HttpOnly`: no
- `SameSite`: `Lax`
- `Path`: `/`

`init=1` is not authentication and must not authorize access to protected data.

Auth session cookie:

- Name: `sid`
- Purpose: authenticated API session
- JavaScript-readable: no
- `HttpOnly`: yes
- `Secure`: yes outside local development
- `SameSite`: `Lax`
- `Path`: `/`

Future SSO integration, if added later, should have a separate contract and migration plan. It is not part of this OpenAPI spec.

## OpenAPI/Ogen Notes

- Keep operation IDs stable; generated Go interfaces depend on them.
- Prefer explicit schemas over anonymous inline objects.
- Avoid broad `oneOf`/`anyOf` unless there is a concrete need.
- Keep S3 details out of the client contract. The API returns stable media URLs or a missing-media response.
- Run `ogen` generation as a contract check once the Go server module exists.
