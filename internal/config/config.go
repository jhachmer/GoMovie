// Package config initialises config variables used in application
package config

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

// Config struct holds fields set by environment variables
type Config struct {
	Addr       string
	OmdbApiKey string
}

var Envs = initConfig()

func initConfig() Config {
	return Config{
		Addr:       GetEnv("ADDR", ":8080"),
		OmdbApiKey: GetEnv("OMDB_KEY", ""),
	}
}

// GetEnv retrieves environment variable with name `key`
// if not present will use fallback value
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
