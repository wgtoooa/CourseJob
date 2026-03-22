package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
}

func MustLoad() Config {
	_ = godotenv.Load()

	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8080"
	}

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		getEnv("DATABASE_USER"),
		getEnv("DATABASE_PASSWORD"),
		getEnv("DATABASE_HOST"),
		getEnv("DATABASE_PORT"),
		getEnv("DATABASE_NAME"),
	)

	return Config{
		HTTPAddr:    httpAddr,
		DatabaseURL: dbURL,
	}
}

func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("env %s is required", key)
	}
	return val
}
