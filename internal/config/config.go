// Package config initialises config variables used in application
package config

import (
	"fmt"
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
	Valid      bool
}

var Envs = initConfig()

func initConfig() Config {
	var valid bool = true
	addr, err := GetEnv("ADDR", ":8080")
	if err != nil {
		valid = false
	}
	omdbKey, err := GetEnv("OMDB_KEY", "")
	if err != nil {
		valid = false
	}
	jwtKey, err := GetEnv("GOLIST_JWT", "")
	if err != nil {
		valid = false
	}
	adminName, err := GetEnv("ADMIN_NAME", "")
	if err != nil {
		valid = false
	}
	adminPw, err := GetEnv("ADMIN_PW", "")
	if err != nil {
		valid = false
	}
	return Config{
		Addr:       addr,
		OmdbApiKey: omdbKey,
		JwtKey:     jwtKey,
		AdminName:  adminName,
		AdminPW:    adminPw,
		Valid:      valid,
	}
}

// GetEnv retrieves environment variable with name `key`
// if not present will use fallback value
func GetEnv(key, fallback string) (string, error) {
	if value, ok := os.LookupEnv(key); ok {
		return value, nil
	}
	if fallback == "" {
		log.Printf("a value (%v) is missing from config!\n", key)
		return "", fmt.Errorf("a value (%v) is missing from config", key)
	}
	return fallback, nil
}
