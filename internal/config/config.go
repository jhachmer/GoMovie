// Package config initialises config variables used in application
package config

import (
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

// Config struct holds fields set by environment variables
type Config struct {
	Addr       string
	OmdbApiKey string
	JwtKey     string
	AdminName  string
	AdminPW    string
}

var Envs = initConfig()

func initConfig() Config {
	return Config{
		Addr:       GetEnv("ADDR", ":8080"),
		OmdbApiKey: GetEnv("OMDB_KEY", ""),
		JwtKey:     GetEnv("GOLIST_JWT", ""),
		AdminName:  GetEnv("ADMIN_NAME", ""),
		AdminPW:    GetEnv("ADMIN_PW", ""),
	}
}

// GetEnv retrieves environment variable with name `key`
// if not present will use fallback value
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	if fallback == "" {
		log.Fatalf("a value (%v) is missing from config!", key)
		os.Exit(1)
	}
	return fallback
}
