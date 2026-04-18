package config

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/ilyakaznacheev/cleanenv"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalYAML(unmarshal func(any) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	v, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	d.Duration = v
	return nil
}

func (d *Duration) UnmarshalText(text []byte) error {
	v, err := time.ParseDuration(string(text))
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", string(text), err)
	}
	d.Duration = v
	return nil
}

type Config struct {
	Query    QueryConfig     `yaml:"query"`
	Postgres postgres.Config `yaml:"postgres"`
	Storage  StorageConfig   `yaml:"storage"`
}

type QueryConfig struct {
	HTTPAddr     string   `yaml:"http_addr" env:"QUERY_HTTP_ADDR" env-default:"0.0.0.0:8080"`
	ReadTimeout  Duration `yaml:"read_timeout" env:"QUERY_READ_TIMEOUT" env-default:"5s"`
	WriteTimeout Duration `yaml:"write_timeout" env:"QUERY_WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout  Duration `yaml:"idle_timeout" env:"QUERY_IDLE_TIMEOUT" env-default:"120s"`
	ShortURLBase string   `yaml:"short_url_base" env:"QUERY_SHORT_URL_BASE" env-default:"http://localhost:8080"`
}

type StorageConfig struct {
	Type string `yaml:"type" env:"STORAGE_TYPE" env-default:"memory"`
}

func Load() (*Config, error) {
	var cfg Config

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		candidates := []string{
			"./configs/config.yaml",
			"configs/config.yaml",
			"../configs/config.yaml",
			"../../configs/config.yaml",
		}
		for _, path := range candidates {
			if _, err := os.Stat(path); err == nil {
				configPath = path
				break
			}
		}
	}

	if configPath != "" {
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			return nil, fmt.Errorf("read config file %q: %w", configPath, err)
		}
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("read env: %w", err)
	}

	return &cfg, nil
}
