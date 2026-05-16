package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Floors   int    `json:"Floors" env:"FLOORS" env-required:"true"`
	Monsters int    `json:"Monsters" env:"MONSTERS" env-required:"true"`
	OpenAt   string `json:"OpenAt" env:"OPENT_AT" env-required:"true"`
	Duration int    `json:"Duration" env:"DURATION" env-required:"true"`
}

func MustLoad(path string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		log.Fatalf("cannot read env: %s", err)
	}
	return cfg
}
