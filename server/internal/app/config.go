package app

import (
	"context"
	"os"
	"time"

	"github.com/dkoshenkov/packages-go/configx"
	"github.com/dkoshenkov/packages-go/middlewarex/httpx"
)

type Config struct {
	ServiceName string         `cfgx:"service_name,default=sport-api" env:"SERVICE_NAME"`
	Env         string         `cfgx:"env,default=dev" env:"ENV"`
	Log         LogConfig      `cfgx:"log"`
	HTTP        httpx.Config   `cfgx:"http"`
	Server      ServerConfig   `cfgx:"server"`
	Database    DatabaseConfig `cfgx:"database"`
	Media       MediaConfig    `cfgx:"media"`
}

type LogConfig struct {
	Level  string `cfgx:"level,default=info" env:"LOG_LEVEL"`
	Pretty bool   `cfgx:"pretty,default=false" env:"LOG_PRETTY"`
}

type ServerConfig struct {
	Addr              string        `cfgx:"addr,default=:8080" env:"SPORT_API_ADDR"`
	ReadHeaderTimeout time.Duration `cfgx:"read_header_timeout,default=5s" env:"HTTP_READ_HEADER_TIMEOUT"`
	ReadTimeout       time.Duration `cfgx:"read_timeout,default=10s" env:"HTTP_READ_TIMEOUT"`
	WriteTimeout      time.Duration `cfgx:"write_timeout,default=30s" env:"HTTP_WRITE_TIMEOUT"`
	IdleTimeout       time.Duration `cfgx:"idle_timeout,default=60s" env:"HTTP_IDLE_TIMEOUT"`
}

type DatabaseConfig struct {
	URL             string        `cfgx:"url,default=postgres://sport:sport@localhost:5432/sport?sslmode=disable" env:"DATABASE_URL"`
	RunMigrations   bool          `cfgx:"run_migrations,default=true" env:"DATABASE_RUN_MIGRATIONS"`
	MaxOpenConns    int32         `cfgx:"max_open_conns,default=10" env:"DATABASE_MAX_OPEN_CONNS"`
	MinIdleConns    int32         `cfgx:"min_idle_conns,default=1" env:"DATABASE_MIN_IDLE_CONNS"`
	MaxConnLifetime time.Duration `cfgx:"max_conn_lifetime,default=1h" env:"DATABASE_MAX_CONN_LIFETIME"`
}

type MediaConfig struct {
	MainDomain string `cfgx:"main_domain,default=example.com" env:"SPORT_MAIN_DOMAIN"`
	BaseURL    string `cfgx:"base_url,optional" env:"EXERCISE_MEDIA_BASE_URL"`
	Manifest   string `cfgx:"manifest,optional" env:"EXERCISE_MEDIA_MANIFEST"`
}

func LoadConfig(ctx context.Context) (Config, error) {
	var cfg Config
	profile := os.Getenv("ENV")
	if profile == "" {
		profile = "dev"
	}
	err := configx.Load(ctx, &cfg,
		configx.ParseFlags(),
		configx.WithProfile(profile),
		configx.WithResolveMode(configx.OverlayDefaultLow),
	)
	if cfg.Media.BaseURL == "" && cfg.Media.MainDomain != "" {
		cfg.Media.BaseURL = "https://media." + cfg.Media.MainDomain
	}
	return cfg, err
}
