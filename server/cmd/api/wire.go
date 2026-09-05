//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/goforj/wire"
	"github.com/jackc/pgx/v5/pgxpool"

	"sport/server/internal/app"
	"sport/server/internal/exercises"
)

func newCatalog(ctx context.Context, cfg app.Config, pool *pgxpool.Pool) (exercises.Repository, error) {
	return exercises.NewPostgresCatalog(ctx, pool, cfg.Exercises.DatasetDir)
}

var applicationSet = wire.NewSet(
	app.LoadConfig,
	newLogger,
	app.NewPostgresPool,
	newCatalog,
	app.NewPostgresStore,
	wire.Bind(new(app.Store), new(*app.PostgresStore)),
	app.NewHandler,
	app.NewSecurity,
	newHTTPServer,
	wire.Struct(new(application), "*"),
)

func initializeApplication(ctx context.Context) (*application, func(), error) {
	wire.Build(applicationSet)
	return nil, nil, nil
}
