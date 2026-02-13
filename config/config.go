package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// ClientConfig holds all client settings.
type ClientConfig struct {
	Server struct {
		Addr  string `toml:"addr"`
		Proxy string `toml:"proxy"`
	} `toml:"server"`
	Dev struct {
		Lobby string `toml:"lobby"`
		Debug bool   `toml:"debug"`
	} `toml:"dev"`
}

// LoadClient decodes a client config from the given TOML file.
func LoadClient(path string) (ClientConfig, error) {
	var cfg ClientConfig
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg, fmt.Errorf("decode %s: %w", path, err)
	}
	return cfg, nil
}
