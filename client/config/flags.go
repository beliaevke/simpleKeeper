package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
)

type ClientFlags struct {
	FlagVersion    string
	FlagAESKeyInfo string `json:"aes_key_info"`
	FlagCryptoCert string `json:"crypto_cert"`
	FlagCryptoKey  string `json:"crypto_key"`
	EnvAESKeyInfo  string `env:"SIMPLEKEEPER_AES_KEY"`
	EnvCryptoCert  string `env:"SIMPLEKEEPER_CRYPTO_CERT"`
	EnvCryptoKey   string `env:"SIMPLEKEEPER_CRYPTO_KEY"`
}

// NewConfig обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func NewConfig() ClientFlags {
	// для случаев, когда в переменных окружения присутствует непустое значение,
	// переопределим их, даже если они были переданы через аргументы командной строки
	cfg := &ClientFlags{}
	if err := env.Parse(cfg); err != nil {
		log.Fatal(err)
	}

	// Регистрируем переменные:
	// как аргумент -ver со значением текущей версии по умолчанию
	flag.StringVar(&cfg.FlagVersion, "ver", "0.1.0", "current version")
	// как аргумент -ak со значением локального каталога по умолчанию
	flag.StringVar(&cfg.FlagAESKeyInfo, "ak", "", "path to AES key")
	// регистрируем переменную FlagCryptoCert
	// как аргумент -ck со значением локального каталога по умолчанию
	flag.StringVar(&cfg.FlagCryptoCert, "ck", "", "path to certificate")
	// регистрируем переменную FlagCryptoKey
	// как аргумент -pk со значением локального каталога по умолчанию
	flag.StringVar(&cfg.FlagCryptoKey, "pk", "", "path to private key")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()

	// для случаев, когда в переменной окружения присутствует непустое значение,
	// используем его, если значение не было передано через аргумент командной строки
	if cfg.FlagAESKeyInfo == "" && cfg.EnvAESKeyInfo != "" {
		cfg.FlagAESKeyInfo = cfg.EnvAESKeyInfo
	}
	if cfg.FlagCryptoCert == "" && cfg.EnvCryptoCert != "" {
		cfg.FlagCryptoCert = cfg.EnvCryptoCert
	}
	if cfg.FlagCryptoKey == "" && cfg.EnvCryptoKey != "" {
		cfg.FlagCryptoKey = cfg.EnvCryptoKey
	}

	return *cfg
}
