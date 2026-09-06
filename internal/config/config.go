package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr      string
	PostgresURL   string
	MongoURI      string
	MongoDatabase string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	KafkaBrokers  []string
	KafkaTopic    string
}

func Load() Config {
	return Config{
		HTTPAddr:      env("HTTP_ADDR", ":8080"),
		PostgresURL:   env("POSTGRES_URL", "postgres://game:game@localhost:15432/game?sslmode=disable"),
		MongoURI:      env("MONGO_URI", "mongodb://localhost:27018"),
		MongoDatabase: env("MONGO_DATABASE", "game"),
		RedisAddr:     env("REDIS_ADDR", "localhost:16379"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       envInt("REDIS_DB", 0),
		KafkaBrokers:  strings.Split(env("KAFKA_BROKERS", "localhost:19092"), ","),
		KafkaTopic:    env("KAFKA_TOPIC", "game.activity"),
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return value
}
