//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/goforj/wire"

	"sport/server/internal/app"
	"sport/server/internal/exercises"
)

func newCatalog(cfg app.Config) (*exercises.Catalog, error) {
	return exercises.NewCatalog(cfg.Media.BaseURL, cfg.Media.Manifest)
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
