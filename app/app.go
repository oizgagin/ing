package app

/*

import (
	"github.com/oizgagin/ing/pkg/cache"
	"github.com/oizgagin/ing/pkg/cache/redisring"
	dbpkg "github.com/oizgagin/ing/pkg/db"
	"github.com/oizgagin/ing/pkg/db/postgres"
	streampkg "github.com/oizgagin/ing/pkg/stream"
	"github.com/oizgagin/ing/pkg/stream/kafka"
)

type Config struct {
	DB    postgres.Config `toml:"db"`
	Kafka kafka.Config    `toml:"kafka"`
	Cache redisring.Cache `toml:"cache"`
	Server s
}

type AppConfig struct {
	ListenAddr string `toml:"listen_addr"`
	MetricsAddr string `toml:"metrics_addr"`
	LogLevel string `toml:"app_config"`
}



type App struct {
	server *server.Server
	db         dbpkg.DB

	stream     streampkg.Stream
	eventCache cache.EventInfoCache
}

func NewApp(cfg Config) (*App, error) {


	db, err := postgres.NewDB(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("could not create db: %w", err)
	}

	stream := kafka.NewStream(



	return nil
}
*/
