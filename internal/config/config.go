package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Connection struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	User    string `yaml:"user"`
	KeyFile string `yaml:"key_file,omitempty"`
}

type Config struct {
	Connections map[string]Connection `yaml:"connections"`
}

func defaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".sftpgo.yml")
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = defaultPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Connections: make(map[string]Connection)}, err
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Connections == nil {
		cfg.Connections = make(map[string]Connection)
	}

	return &cfg, nil
}

func Save(path string, cfg *Config) error {
	if path == "" {
		path = defaultPath()
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
