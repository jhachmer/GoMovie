// Package config initialises config variables used in application
package config

import (
	"fmt"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type DBConfig interface {
	ConnectionString() string
}

type SQLiteConfig struct {
	Path string
}

func (c SQLiteConfig) ConnectionString() string {
	return fmt.Sprintf("file:%s", c.Path)
}

// Config struct holds fields set by environment variables
type Config struct {
	Addr       string
	OmdbApiKey string
	JwtKey     string
	AdminName  string
	AdminPW    string
	DbType     string
	DbConfig   DBConfig

	Valid bool
}

var Envs = initConfig()

func initConfig() Config {
	var valid = true
	var dbConfig DBConfig
	addr, err := GetEnv("ADDR", ":8080")
	if err != nil {
		valid = false
	}
	omdbKey, err := GetEnv("OMDB_KEY", "")
	if err != nil || omdbKey == "" {
		valid = false
	}
	jwtKey, err := GetEnv("GOMOVIE_JWT", "uns3cure_jwt")
	if err != nil {
		valid = false
	}
	adminName, err := GetEnv("ADMIN_NAME", "")
	if err != nil || adminName == "" {
		valid = false
	}
	adminPw, err := GetEnv("ADMIN_PW", "")
	if err != nil || adminPw == "" {
		valid = false
	}
	dbType, err := GetEnv("DB_TYPE", "")
	if err != nil {
		dbConfig = SQLiteConfig{
			Path: "./gomovie.sqlite",
		}
	} else {
		switch dbType {
		case "sqlite3":
			sqlitePath, err := GetEnv("SQLITE_PATH", "./gomovie.sqlite")
			if err != nil {
				valid = false
			}
			dbConfig = SQLiteConfig{
				Path: sqlitePath,
			}
		}
	}
	return Config{
		Addr:       addr,
		OmdbApiKey: omdbKey,
		JwtKey:     jwtKey,
		AdminName:  adminName,
		AdminPW:    adminPw,
		DbConfig:   dbConfig,
		DbType:     dbType,
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
		return "", fmt.Errorf("a value (%v) is missing from config", key)
	}
	return fallback, nil
}
