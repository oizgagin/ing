package app

import (
	"errors"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/oizgagin/ing/app/rsvphandler"
	"github.com/oizgagin/ing/app/server"
	"github.com/oizgagin/ing/pkg/cache"
	"github.com/oizgagin/ing/pkg/cache/redisring"
	"github.com/oizgagin/ing/pkg/db"
	"github.com/oizgagin/ing/pkg/db/postgres"
	"github.com/oizgagin/ing/pkg/stream"
	"github.com/oizgagin/ing/pkg/stream/kafka"
)

type AppConfig struct {
	Output     string `toml:"output"`
	LogLevel   string `toml:"log_level"`
	MetricAddr string `toml:"metric_addr"`
}

type Config struct {
	App         AppConfig          `toml:"app"`
	Kafka       kafka.Config       `toml:"kafka"`
	Postgres    postgres.Config    `toml:"postgres"`
	RedisRing   redisring.Config   `toml:"redis-ring"`
	RSVPHandler rsvphandler.Config `toml:"rsvp-handler"`
	Server      server.Config      `toml:"server"`
}

type App struct {
	l *zap.Logger

	stream stream.Stream
	db     db.DB
	cache  cache.EventInfoCache

	handler       *rsvphandler.Handler
	server        *server.Server
	metricsServer *metricsServer
}

func NewApp(cfg Config) (*App, error) {

	l, err := buildLogger(cfg.App.LogLevel, cfg.App.Output)
	if err != nil {
		return nil, fmt.Errorf("could not init logging: %w", err)
	}

	db, err := postgres.NewDB(cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("could not create db: %w", err)
	}

	cache, err := redisring.NewCache(cfg.RedisRing)
	if err != nil {
		return nil, fmt.Errorf("could not create cache: %w", err)
	}

	stream := kafka.NewStream(cfg.Kafka, l)

	handler := rsvphandler.NewHandler(cfg.RSVPHandler, l, stream, db)

	server, err := server.NewServer(cfg.Server, l, db, cache)
	if err != nil {
		return nil, fmt.Errorf("could not create server: %w", err)
	}

	metricsServer := newMetricsServer(l, cfg.App.MetricAddr)

	app := App{
		l:             l,
		stream:        stream,
		db:            db,
		cache:         cache,
		handler:       handler,
		server:        server,
		metricsServer: metricsServer,
	}

	return &app, nil
}

func (app *App) Close() error {
	app.l.Info("closing application")

	app.handler.Stop()

	var errs []error

	errs = append(errs, app.server.Close())
	errs = append(errs, app.db.Close())
	errs = append(errs, app.stream.Close())
	errs = append(errs, app.cache.Close())
	errs = append(errs, app.metricsServer.Close())

	err := errors.Join(errs...)

	if err != nil {
		app.l.Error("close application error", zap.Error(err))
		return err
	}

	app.l.Info("application closed")
	return nil
}

func buildLogger(level, output string) (*zap.Logger, error) {
	if level != "debug" && level != "info" && level != "error" {
		return nil, fmt.Errorf(`invalid log-level %q, must be in ("debug", "info" or "error")`, level)
	}
	if output != "stderr" && output != "stdout" {
		return nil, fmt.Errorf(`invalid log output %q, must be in ("stderr" or "stdout")`, output)
	}

	out := os.Stderr
	if output == "stdout" {
		out = os.Stdout
	}

	atom := zap.NewAtomicLevel()

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.Lock(out),
		atom,
	))

	switch level {
	case "debug":
		atom.SetLevel(zap.DebugLevel)
	case "info":
		atom.SetLevel(zap.InfoLevel)
	case "error":
		atom.SetLevel(zap.ErrorLevel)
	}

	return logger, nil
}
