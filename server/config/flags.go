package config

import (
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env"
)

type ServerFlags struct {
	FlagDatabaseURI string
	FlagTLSCert     string `json:"tls_cert"`
	FlagTLSKey      string `json:"tls_key"`
	EnvDatabaseURI  string `env:"DATABASE_URI"`
	EnvTLSCert      string `env:"SIMPLEKEEPER_TLS_CERT"`
	EnvTLSKey       string `env:"SIMPLEKEEPER_TLS_KEY"`
	DefaultTimeout  time.Duration
}

// NewConfig обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func NewConfig() ServerFlags {
	// для случаев, когда в переменных окружения присутствует непустое значение,
	// переопределим их, даже если они были переданы через аргументы командной строки
	cfg := &ServerFlags{}
	if err := env.Parse(cfg); err != nil {
		log.Fatal(err)
	}

	// Регистрируем переменные:
	// Строка с адресом подключения к БД должна получаться из переменной окружения DATABASE_DSN или флага командной строки -d
	flag.StringVar(&cfg.FlagDatabaseURI, "d", "postgres://postgres:pos111@localhost:5432/postgres?sslmode=disable", "Database URI")
	// регистрируем переменную FlagTLSCert
	// как аргумент -tc со значением локального каталога по умолчанию
	flag.StringVar(&cfg.FlagTLSCert, "tc", "", "path to TLS certificate")
	// регистрируем переменную FlagTLSKey
	// как аргумент -tk со значением локального каталога по умолчанию
	flag.StringVar(&cfg.FlagTLSKey, "tk", "", "path to TLS key")
	// Продолжительность таймаутов
	flag.DurationVar(&cfg.DefaultTimeout, "dt", 30*time.Second, "Default timeout duration")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()

	// для случаев, когда в переменной окружения присутствует непустое значение,
	// используем его, если значение не было передано через аргумент командной строки
	if cfg.FlagDatabaseURI == "" && cfg.EnvDatabaseURI != "" {
		cfg.FlagDatabaseURI = cfg.EnvDatabaseURI
	}
	if cfg.FlagTLSCert == "" && cfg.EnvTLSCert != "" {
		cfg.FlagTLSCert = cfg.EnvTLSCert
	}
	if cfg.FlagTLSKey == "" && cfg.EnvTLSKey != "" {
		cfg.FlagTLSKey = cfg.EnvTLSKey
	}

	return *cfg
}
