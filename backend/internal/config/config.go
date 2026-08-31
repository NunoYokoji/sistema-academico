// Package config carrega a configuração da aplicação a partir de variáveis
// de ambiente — as mesmas que o ConfigMap e o Secret do Kubernetes injetam.
package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port        string
	DatabaseDSN string
	RedisAddr   string
	CacheTTL    time.Duration
	Seed        bool
}

// Load monta a configuração. Não há valor padrão para a senha do banco:
// ela vem sempre do Secret.
func Load() Config {
	dbHost := env("DB_HOST", "localhost")
	dbPort := env("DB_PORT", "5432")
	dbName := env("DB_NAME", "sistema_academico")
	dbUser := env("DB_USER", "postgres")
	dbPass := os.Getenv("DB_PASSWORD")

	return Config{
		Port: env("PORT", "8080"),
		DatabaseDSN: fmt.Sprintf(
			"postgres://%s:%s@%s/%s?sslmode=disable",
			dbUser, dbPass, net.JoinHostPort(dbHost, dbPort), dbName,
		),
		RedisAddr: net.JoinHostPort(env("REDIS_HOST", "localhost"), env("REDIS_PORT", "6379")),
		CacheTTL:  time.Duration(envInt("CACHE_TTL_SECONDS", 60)) * time.Second,
		Seed:      env("SEED_ON_EMPTY", "true") == "true",
	}
}

func env(chave, padrao string) string {
	if v := os.Getenv(chave); v != "" {
		return v
	}
	return padrao
}

func envInt(chave string, padrao int) int {
	if v, err := strconv.Atoi(os.Getenv(chave)); err == nil && v > 0 {
		return v
	}
	return padrao
}
