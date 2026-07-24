package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string // Database Utama (db_unipack)
	DBNuxtName string // Database Baru (db_nuxt)
}

func LoadConfig() *Config {
	// Load file .env jika ada
	_ = godotenv.Load()

	return &Config{
		DBUser:     getEnv("DB_USERNAME", "root"),
		DBPassword: getEnv("DB_PASSWORD", "admin"),
		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBName:     getEnv("DB_NAME", "db_unipack_lokal"),
		DBNuxtName: getEnv("DB_NUXT_NAME", "db_nuxt"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}