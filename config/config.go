package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	AppEnv     string
	JwtSecret  string
	DBUser     string
	DBPassword string
	DBName     string
	DBHost     string
	DBPort     string
	DBSSLMode  string
	Port       string
)

// LoadConfig load environment variables into memory
func LoadConfig() {

	if err := godotenv.Load(); err != nil {
		log.Println("Can't find .env file or error loading it")
	}

	AppEnv = getEnv("APP_ENV", "development")
	JwtSecret = getEnv("JWT_SECRET", "")
	DBUser = getEnv("DB_USER", "")
	DBPassword = getEnv("DB_PASSWORD", "")
	DBName = getEnv("DB_NAME", "")
	DBHost = getEnv("DB_HOST", "")
	DBPort = getEnv("DB_PORT", "5432")
	Port = getEnv("PORT", "8000")
	DBSSLMode = getEnv("DB_SSLMODE", "disable")

	//Some validation
	if JwtSecret == "" {
		log.Fatal("Missing JWT_SECRET value in .env file")
	}
	if DBUser == "" || DBPassword == "" || DBName == "" {
		log.Fatal("Missing database details and credentials in .env file")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
