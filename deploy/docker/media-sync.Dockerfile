FROM golang:1.25.6-alpine AS build

WORKDIR /src/server

RUN apk add --no-cache ca-certificates git

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/sync-exercise-media ./cmd/sync-exercise-media

FROM alpine:3.22

RUN apk add --no-cache ca-certificates git

WORKDIR /app

COPY --from=build /out/sync-exercise-media /app/sync-exercise-media
COPY server/internal/exercises/ru_overrides.json /app/internal/exercises/ru_overrides.json

ENTRYPOINT ["/app/sync-exercise-media"]
