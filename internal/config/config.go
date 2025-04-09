package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBDSN     string
	JWTSecret string
	Port      string
}

func LoadConfig() Config {
	dbHost := getEnv("DATABASE_HOST", "db")
	dbPort := getEnv("DATABASE_PORT", "5432")
	dbUser := getEnv("DATABASE_USER", "postgres")
	dbPass := getEnv("DATABASE_PASSWORD", "postgres")
	dbName := getEnv("DATABASE_NAME", "pvz")

	return Config{
		DBDSN:     fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName),
		JWTSecret: getEnv("JWT_SECRET", "secret"),
		Port:      getEnv("SERVER_PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
