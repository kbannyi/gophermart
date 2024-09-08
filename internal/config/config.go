package config

import (
	"errors"
	"flag"
	"os"
)

type Config struct {
	RunAddr     string
	DatabaseURI string
}

func ParseConfig() (Config, error) {
	cfg := Config{}
	flag.StringVar(&cfg.RunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "db connection string")
	flag.Parse()

	if env := os.Getenv("RUN_ADDRESS"); env != "" {
		cfg.RunAddr = env
	}
	if env := os.Getenv("DATABASE_URI"); env != "" {
		cfg.DatabaseURI = env
	}

	if cfg.DatabaseURI == "" {
		return cfg, errors.New("db connection string is required")
	}

	return cfg, nil
}
