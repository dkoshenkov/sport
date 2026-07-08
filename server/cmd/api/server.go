package main

import (
	"net/http"

	"github.com/rs/zerolog"

	"sport/server/internal/api"
	"sport/server/internal/app"
	"sport/server/pkg/httpserver"
	"sport/server/pkg/logging"
)

type application struct {
	Logger     zerolog.Logger
	HTTPServer *http.Server
}

func newLogger(cfg app.Config) (zerolog.Logger, error) {
	return logging.New(logging.Config{
		ServiceName: cfg.ServiceName,
		Env:         cfg.Env,
		Level:       cfg.Log.Level,
		Pretty:      cfg.Log.Pretty || cfg.HTTP.PrettyLogs,
	})
}

func newHTTPServer(cfg app.Config, logger zerolog.Logger, handler *app.Handler, security *app.Security) (*http.Server, error) {
	ogenServer, err := api.NewServer(handler, security, api.WithErrorHandler(app.ErrorHandler))
	if err != nil {
		return nil, err
	}

	return httpserver.New(ogenServer, httpserver.Config{
		Addr:              cfg.Server.Addr,
		RequestIDHeader:   cfg.HTTP.RequestIDHeader,
		Timeout:           cfg.HTTP.Timeout,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
	}, logger), nil
}
