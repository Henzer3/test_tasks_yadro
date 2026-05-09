package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel string `env:"LOG_LEVEL" env-default:"DEBUG"`
	Path     string `env:"PATH" env-default:"/etc/resolv.conf"`
	GRPC     GRPCConfig
}

type GRPCConfig struct {
	Address              string `env:"DNS_ADDRESS" env-required:"true"`
	GracefulShutdownTime int    `env:"GRACEFULSHUTDOWN" env-default:"2"`
}

func MustLoad() Config {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("cannot read env: %s", err)
	}
	return cfg
}
