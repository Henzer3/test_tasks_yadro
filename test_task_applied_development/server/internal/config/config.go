package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel  string `env:"LOG_LEVEL" env-default:"DEBUG"`
	DBAddress string `env:"DB_ADDRESS" env-required:"true"`
	Rest      RestConfig
}

type RestConfig struct {
	Address              string        `env:"APP_ADDRESS" env-required:"true"`
	GracefulShutdownTime int           `env:"GRACEFULSHUTDOWN" env-default:"3"`
	Timeout              time.Duration `yaml:"timeout" env:"LOG_TIMEOUT" env-default:"5s"`
}

func MustLoad() Config {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("cannot read env: %s", err)
	}
	return cfg
}
