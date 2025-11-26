// Package config initialises config variables used in application
package config

import (
	"fmt"
	"log"
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

type PostgresConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

// TODO: remove sslmode=disable
func (c PostgresConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", c.Host, c.Port, c.Username, c.Password, c.Database)
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
	dbType, err := GetEnv("DB_TYPE", "")
	if err != nil {
		dbConfig = SQLiteConfig{
			Path: "./gomovie.sqlite",
		}
	}
	switch dbType {
	case "postgres":
		host, err := GetEnv("POSTGRES_HOST", "localhost")
		if err != nil {
			valid = false
		}
		port, err := GetEnv("POSTGRES_PORT", "5432")
		if err != nil {
			valid = false
		}
		user, err := GetEnv("POSTGRES_USER", "postgres")
		if err != nil {
			valid = false
		}
		password, err := GetEnv("POSTGRES_PASSWORD", "postgres")
		if err != nil {
			valid = false
		}
		database, err := GetEnv("POSTGRES_DB", "postgres")
		if err != nil {
			valid = false
		}
		dbConfig = &PostgresConfig{
			Host:     host,
			Port:     port,
			Username: user,
			Password: password,
			Database: database,
		}
	case "sqlite":
		sqlitePath, err := GetEnv("SQLITE_PATH", "./gomovie.sqlite")
		if err != nil {
			valid = false
		}
		dbConfig = SQLiteConfig{
			Path: sqlitePath,
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
		log.Printf("a value (%v) is missing from config!\n", key)
		return "", fmt.Errorf("a value (%v) is missing from config", key)
	}
	return fallback, nil
}
