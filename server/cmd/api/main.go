package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dkoshenkov/packages-go/logx"
	"github.com/dkoshenkov/packages-go/middlewarex"
	"github.com/dkoshenkov/packages-go/middlewarex/httpx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"sport/server/internal/api"
	"sport/server/internal/app"
	"sport/server/internal/exercises"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "api server stopped: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := app.LoadConfig(ctx)
	if err != nil {
		return err
	}
	logger, err := newLogger(cfg)
	if err != nil {
		return err
	}
	ctx = logx.WithContext(ctx, logger)

	poolCfg, err := pgxpool.ParseConfig(cfg.Database.URL)
	if err != nil {
		return fmt.Errorf("parse database url: %w", err)
	}
	poolCfg.MaxConns = cfg.Database.MaxOpenConns
	poolCfg.MinConns = cfg.Database.MinIdleConns
	poolCfg.MaxConnLifetime = cfg.Database.MaxConnLifetime
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}
	if cfg.Database.RunMigrations {
		if err := app.RunMigrations(ctx, pool); err != nil {
			return err
		}
	}

	catalog, err := exercises.NewCatalog(cfg.Media.BaseURL, cfg.Media.Manifest)
	if err != nil {
		return err
	}
	store := app.NewPostgresStore(pool)
	handler := app.NewHandler(store, catalog)
	security := app.NewSecurity(store)
	ogenServer, err := api.NewServer(handler, security, api.WithErrorHandler(app.ErrorHandler))
	if err != nil {
		return err
	}

	httpHandler := httpx.Wrap(
		ogenServer,
		injectLogger(logger),
		httpx.RequestID(httpx.WithRequestIDHeader(cfg.HTTP.RequestIDHeader)),
		middlewarex.Recovery[httpx.Exchange, struct{}](logMiddleware(logger)),
		middlewarex.Timeout[httpx.Exchange, struct{}](cfg.HTTP.Timeout),
		httpx.Logging(logMiddleware(logger)),
	)

	httpServer := &http.Server{
		Addr:              cfg.Server.Addr,
		Handler:           httpHandler,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
	}

	runCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	errc := make(chan error, 1)
	go func() {
		logx.Info(ctx).Str("addr", cfg.Server.Addr).Msg("api server listening")
		errc <- httpServer.ListenAndServe()
	}()

	select {
	case <-runCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		shutdownCtx = logx.WithContext(shutdownCtx, logger)
		return httpServer.Shutdown(shutdownCtx)
	case err := <-errc:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func newLogger(cfg app.Config) (zerolog.Logger, error) {
	opts := []logx.Option{
		logx.WithLevelText(cfg.Log.Level),
		logx.WithField("env", cfg.Env),
	}
	if cfg.Log.Pretty || cfg.HTTP.PrettyLogs {
		opts = append(opts, logx.WithPretty())
	}
	return logx.New(cfg.ServiceName, opts...)
}

func injectLogger(logger zerolog.Logger) httpx.Middleware {
	return func(next httpx.Handler) httpx.Handler {
		return func(ctx context.Context, exchange httpx.Exchange) (struct{}, error) {
			ctx = logx.WithContext(ctx, logger)
			exchange.Request = exchange.Request.WithContext(ctx)
			return next(ctx, exchange)
		}
	}
}

func logMiddleware(logger zerolog.Logger) middlewarex.Logger {
	return middlewarex.LoggerFunc(func(ctx context.Context, event middlewarex.Event) {
		contextLogger := logx.FromContext(ctx)
		entry := contextLogger.WithLevel(parseLevel(event.Level)).
			Str("name", event.Name).
			Dur("duration", event.Duration)
		if event.RequestID != "" {
			entry = entry.Str("request_id", event.RequestID)
		}
		if event.Subject != "" {
			entry = entry.Str("subject", event.Subject)
		}
		for key, value := range event.Fields {
			entry = entry.Interface(key, value)
		}
		if event.Err != nil {
			entry = entry.Err(event.Err)
		}
		if entry == nil {
			entry = logger.WithLevel(parseLevel(event.Level))
		}
		entry.Msg(event.Message)
	})
}

func parseLevel(level string) zerolog.Level {
	parsed, err := zerolog.ParseLevel(level)
	if err != nil {
		return zerolog.InfoLevel
	}
	return parsed
}
