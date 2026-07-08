package logging

import (
	"github.com/dkoshenkov/packages-go/logx"
	"github.com/rs/zerolog"
)

type Config struct {
	ServiceName string
	Env         string
	Level       string
	Pretty      bool
}

func New(cfg Config) (zerolog.Logger, error) {
	opts := []logx.Option{
		logx.WithLevelText(cfg.Level),
		logx.WithField("env", cfg.Env),
	}
	if cfg.Pretty {
		opts = append(opts, logx.WithPretty())
	}
	return logx.New(cfg.ServiceName, opts...)
}
