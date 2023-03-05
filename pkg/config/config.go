package config

import (
	"fmt"
	"os"
	"regexp"

	"github.com/BurntSushi/toml"
)

var (
	envRe = regexp.MustCompile(`{%\s+ENV:(\S+)\s+%}`)
)

func ParseFile(filename string, v any) error {
	raw, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read config from %v: %v", filename, err)
	}

	raw = envRe.ReplaceAllFunc(raw, func(env []byte) []byte {
		return []byte(os.Getenv(string(envRe.FindSubmatch(env)[1])))
	})

	if err := toml.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("could not parse config from %v: %v", filename, err)
	}
	return nil
}
