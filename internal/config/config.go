package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr       string
	DatabaseURL    string
	JWTSecret      string
	JWTTTLMinutes  int
	RefreshTTLMins int
	UIDHMACSecret  string
	RequireHTTPS   bool
	TrustProxyTLS  bool
}

func MustLoad() Config {
	_ = godotenv.Load()

	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8080"
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			getEnv("DATABASE_USER"),
			getEnv("DATABASE_PASSWORD"),
			getEnv("DATABASE_HOST"),
			getEnv("DATABASE_PORT"),
			getEnv("DATABASE_NAME"),
		)
	}

	shaKey := getEnvOrDefault("SHA_KEY", "dscjsbnjkwe3")
	jwtSecret := getEnvOrDefault("JWT_SECRET", shaKey)
	uidHMACSecret := getEnvOrDefault("UID_HMAC_SECRET", "coursejob-card-uid-v1")

	jwtTTLMinutes := 120
	if raw := os.Getenv("JWT_TTL_MINUTES"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			log.Fatalf("env JWT_TTL_MINUTES must be a positive integer")
		}
		jwtTTLMinutes = parsed
	}

	refreshTTLMins := 60 * 24 * 30
	if raw := os.Getenv("REFRESH_TTL_MINUTES"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			log.Fatalf("env REFRESH_TTL_MINUTES must be a positive integer")
		}
		refreshTTLMins = parsed
	}

	requireHTTPS := getEnvBool("REQUIRE_HTTPS", true)
	trustProxyTLS := getEnvBool("TRUST_PROXY_TLS", true)

	return Config{
		HTTPAddr:       httpAddr,
		DatabaseURL:    dbURL,
		JWTSecret:      jwtSecret,
		JWTTTLMinutes:  jwtTTLMinutes,
		RefreshTTLMins: refreshTTLMins,
		UIDHMACSecret:  uidHMACSecret,
		RequireHTTPS:   requireHTTPS,
		TrustProxyTLS:  trustProxyTLS,
	}
}

func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("env %s is required", key)
	}
	return val
}

func getEnvOrDefault(key string, defaultValue string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return defaultValue
	}
	return val
}

func getEnvBool(key string, defaultValue bool) bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if raw == "" {
		return defaultValue
	}

	switch raw {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		log.Fatalf("env %s must be boolean (true/false, 1/0, yes/no)", key)
		return defaultValue
	}
}
