package postgres

import "net/url"

type Config struct {
	Addr   string `toml:"addr"`
	User   string `toml:"user"`
	Pass   string `toml:"pass"`
	DBName string `toml:"dbname"`
}

func (c Config) DSN() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Pass),
		Host:   c.Addr,
		Path:   c.DBName,
	}
	return u.String()
}
