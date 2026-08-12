FROM golang:1.25.6-alpine AS build

WORKDIR /src/server

RUN apk add --no-cache ca-certificates

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.22 AS api

RUN apk add --no-cache ca-certificates && adduser -D -H -u 10001 app

USER app
WORKDIR /app

COPY --from=build /out/api /app/api
COPY exercises-dataset-main /app/exercises-dataset-main

EXPOSE 8080
ENTRYPOINT ["/app/api"]
