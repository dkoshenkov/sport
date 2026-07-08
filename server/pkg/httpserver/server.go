package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/dkoshenkov/packages-go/logx"
	"github.com/dkoshenkov/packages-go/middlewarex"
	"github.com/dkoshenkov/packages-go/middlewarex/httpx"
	"github.com/rs/zerolog"
)

type Config struct {
	Addr              string
	RequestIDHeader   string
	Timeout           time.Duration
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

func New(handler http.Handler, cfg Config, logger zerolog.Logger) *http.Server {
	httpHandler := httpx.Wrap(
		handler,
		InjectLogger(logger),
		httpx.RequestID(httpx.WithRequestIDHeader(cfg.RequestIDHeader)),
		middlewarex.Recovery[httpx.Exchange, struct{}](EventLogger(logger)),
		middlewarex.Timeout[httpx.Exchange, struct{}](cfg.Timeout),
		httpx.Logging(EventLogger(logger)),
	)

	return &http.Server{
		Addr:              cfg.Addr,
		Handler:           httpHandler,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
}

func InjectLogger(logger zerolog.Logger) httpx.Middleware {
	return func(next httpx.Handler) httpx.Handler {
		return func(ctx context.Context, exchange httpx.Exchange) (struct{}, error) {
			ctx = logx.WithContext(ctx, logger)
			exchange.Request = exchange.Request.WithContext(ctx)
			return next(ctx, exchange)
		}
	}
}

func EventLogger(logger zerolog.Logger) middlewarex.Logger {
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
