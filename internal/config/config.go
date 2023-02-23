package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v7"
)

type Config struct {
	Address        string `env:"RUN_ADDRESS"`
	DatabaseURI    string `env:"DATABASE_URI"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`

	Debug bool `env:"DEBUG"`
}

func NewConfig() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.Address, "a", "0.0.0.0:8080", "serve address")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "db connect string")
	flag.StringVar(&cfg.AccrualAddress, "r", "", "cash calculations system address")
	flag.BoolVar(&cfg.Debug, "debug", false, "debug mode")

	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c Config) String() string {
	return fmt.Sprintf(
		"address: '%s', db: '%s', accural: '%s'",
		c.Address,
		c.DatabaseURI,
		c.AccrualAddress,
	)
}
