package config

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/caarlos0/env/v11"
)

type Base struct {
	LogLevel slog.Level `env:"LOG_LEVEL" envDefault:"info"`
}

type Postgres struct {
	Addr     string `env:"ADDR" envDefault:"localhost:5432"`
	Name     string `env:"DB_NAME" envDefault:"billing"`
	User     string `env:"USER" envDefault:"mazafaka"`
	Password string `env:"PASSWORD,required" json:"-"`
}

func (p Postgres) DSN() string {
	return fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=disable", p.User, p.Password, p.Addr, p.Name)
}

type Billing struct {
	Base
	Addr     string   `env:"SERVER_ADDR" envDefault:":8080"`
	Postgres Postgres `envPrefix:"POSTGRES_"`
}

type validator interface {
	Validate() error
}

func Load[T validator]() (T, error) {
	cfg, err := env.ParseAs[T]()
	if err != nil {
		return cfg, fmt.Errorf("failed to parse: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return cfg, fmt.Errorf("failed to validate: %w", err)
	}

	return cfg, nil
}

func (c Billing) Validate() error {
	var errs []error

	if c.Postgres.Addr == "" {
		errs = append(errs, errors.New("POSTGRES_ADDR is empty"))
	}
	if c.Addr == "" {
		errs = append(errs, errors.New("SERVER_ADDR is empty"))
	}
	if c.Postgres.Name == "" {
		errs = append(errs, errors.New("POSTGRES_DB_NAME is empty"))
	}
	if c.Postgres.User == "" {
		errs = append(errs, errors.New("POSTGRES_USER is empty"))
	}

	return errors.Join(errs...)
}
