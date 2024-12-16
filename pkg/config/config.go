package config

import (
	_ "github.com/joho/godotenv/autoload"
	"os"
)

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

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
