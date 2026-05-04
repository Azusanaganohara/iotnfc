package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	Env  string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	JWTSecret             string
	JWTAccessDurationMin  int
	JWTRefreshDurationDay int

	ProvisionSecret        string
	RegisterModeTimeoutMin int
	DeviceAPIKey           string
	DeviceMasterKey        string

	AdminEmail    string
	AdminPassword string

	AllowedOrigins string
}

var instance *Config

func Load() *Config {
	_ = godotenv.Load()
	instance = &Config{
		Port: getEnv("PORT", "8080"),
		Env:  getEnv("APP_ENV", "development"),

		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "iot_ktp"),

		JWTSecret:             getEnv("JWT_SECRET", "change-this-secret-in-production-32chars"),
		JWTAccessDurationMin:  getEnvInt("JWT_ACCESS_DURATION_MIN", 15),
		JWTRefreshDurationDay: getEnvInt("JWT_REFRESH_DURATION_DAYS", 7),

		ProvisionSecret:        getEnv("PROVISION_SECRET", ""),
		RegisterModeTimeoutMin: getEnvInt("REGISTER_MODE_TIMEOUT_MIN", 0),
		DeviceAPIKey:           getEnv("DEVICE_API_KEY", ""),
		DeviceMasterKey:        getEnv("DEVICE_MASTER_KEY", ""),

		AdminEmail:    getEnv("ADMIN_EMAIL", "admin@example.com"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "secret123"),

		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
	}
	return instance
}

func Get() *Config {
	if instance == nil {
		return Load()
	}
	return instance
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}
