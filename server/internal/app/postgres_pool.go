package app

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(ctx context.Context, cfg Config) (*pgxpool.Pool, func(), error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.Database.URL)
	if err != nil {
		return nil, nil, fmt.Errorf("parse database url: %w", err)
	}
	poolCfg.MaxConns = cfg.Database.MaxOpenConns
	poolCfg.MinConns = cfg.Database.MinIdleConns
	poolCfg.MaxConnLifetime = cfg.Database.MaxConnLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("connect postgres: %w", err)
	}
	cleanup := func() {
		pool.Close()
	}

	if err := pool.Ping(ctx); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("ping postgres: %w", err)
	}
	if cfg.Database.RunMigrations {
		if err := RunMigrations(ctx, pool); err != nil {
			cleanup()
			return nil, nil, err
		}
	}

	return pool, cleanup, nil
}
