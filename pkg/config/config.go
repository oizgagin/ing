package config

import (
	"fmt"
	"os"
	"regexp"

	"github.com/BurntSushi/toml"
	"github.com/oizgagin/ing/pkg/stream/kafka"
)

type Config struct {
	Kafka kafka.Config `toml:"kafka"`

	Redis struct {
		Addrs []string `toml:"addrs"`
		User  string   `toml:"user"`
		Pass  string   `toml:"pass"`
	} `toml:"redis"`

	Postgres struct {
		Addr   string `toml:"addr"`
		User   string `toml:"user"`
		Pass   string `toml:"pass"`
		DBName string `toml:"dbname"`
	} `toml:"postgres"`
}

var (
	envRe = regexp.MustCompile(`{%\s+ENV:(\S+)\s+%}`)
)

func ParseFile(filename string) (Config, error) {
	raw, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, fmt.Errorf("could not read config from %v: %v", filename, err)
	}

	raw = envRe.ReplaceAllFunc(raw, func(env []byte) []byte {
		return []byte(os.Getenv(string(envRe.FindSubmatch(env)[1])))
	})

	var c Config
	if err := toml.Unmarshal(raw, &c); err != nil {
		return Config{}, fmt.Errorf("could not parse config from %v: %v", filename, err)
	}

	return c, nil
}
