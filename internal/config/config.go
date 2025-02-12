package config

import (
	"errors"
	"flag"
	"os"
)

type Config struct {
	RunAddr     string
	DatabaseURI string
	AccrualAddr string
}

func ParseConfig() (Config, error) {
	cfg := Config{}
	flag.StringVar(&cfg.RunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "db connection string")
	flag.StringVar(&cfg.AccrualAddr, "r", "", "accrual http addr")
	flag.Parse()

	if env := os.Getenv("RUN_ADDRESS"); env != "" {
		cfg.RunAddr = env
	}
	if env := os.Getenv("DATABASE_URI"); env != "" {
		cfg.DatabaseURI = env
	}
	if env := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); env != "" {
		cfg.AccrualAddr = env
	}

	if cfg.DatabaseURI == "" {
		return cfg, errors.New("db connection string is required")
	}

	if cfg.AccrualAddr == "" {
		return cfg, errors.New("accrual http addr is required")
	}

	return cfg, nil
}
