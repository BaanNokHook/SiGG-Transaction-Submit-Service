package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	// Config -.
	Config struct {
		App
		HTTP
		Log
		RMQ
		Loaffinity
		Redis
	}

	// App -.
	App struct {
		Name    string `env-required:"true" env:"APP_NAME"`
		Version string `env-required:"true" env:"APP_VERSION"`
	}

	// HTTP -.
	HTTP struct {
		Port string `env-required:"true" env:"HTTP_PORT"`
	}

	// Log -.
	Log struct {
		Level string `env-required:"true" env:"LOG_LEVEL"`
	}

	//RMQ -.
	RMQ struct {
		ClientExchange string `env-required:"true" env:"RMQ_RPC_CLIENT"`
		URL            string `env-required:"true" env:"RMQ_URL"`
	}

	//Loaffinity
	Loaffinity struct {
		URL      string `env-required:"true" env:"LOAFFINITY_RPC_URL"`
		Username string `env-required:"true" env:"LOAFFINITY_RPC_USERNAME"`
		Password string `env-required:"true" env:"LOAFFINITY_RPC_PASSWORD"`
	}

	//Redis
	Redis struct {
		Addr     string `env-required:"true" env:"REDIS_ADDR"`
		Password string `env-required:"true" env:"REDIS_PASSWORD"`
		DB       int    `env-required:"true" env:"REDIS_DB"`
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	if _, err := os.Stat(".env"); err == nil {
		err = cleanenv.ReadConfig(".env", cfg)
		if err != nil {
			return nil, fmt.Errorf("config error: %w", err)
		}
	}

	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
