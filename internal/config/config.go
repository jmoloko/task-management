package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config все параметры конфигурации приложения
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Auth     AuthConfig
	Logger   LoggerConfig
}

// ServerConfig настройки HTTP-сервера
type ServerConfig struct {
	Port         int           `yaml:"port"`
	Host         string        `yaml:"host"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
	IdleTimeout  time.Duration `yaml:"idleTimeout"`
}

// DatabaseConfig настройки подключения к базе данных
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

// RedisConfig настройки подключения к Redis
type RedisConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	DB   int    `yaml:"db"`
}

// AuthConfig настройки аутентификации
type AuthConfig struct {
	SigningKey string        `yaml:"signingKey"`
	TokenTTL   time.Duration `yaml:"tokenTTL"`
}

// LoggerConfig настройки логирования
type LoggerConfig struct {
	Level       string `env:"LOG_LEVEL" envDefault:"info"`
	File        string `env:"LOG_FILE" envDefault:""`
	Format      string `env:"LOG_FORMAT" envDefault:"text"`
	ServiceName string `env:"SERVICE_NAME" envDefault:"task-manager"`
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
}

// LoadConfig загружает конфигурацию из yaml файла
func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return &cfg, nil
}

// Load загружает и возвращает конфигурацию из переменных окружения
func Load() (*Config, error) {
	// Загрузка переменных окружения из .env файла, если он существует
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Port:         getIntEnv("SERVER_PORT", 8080),
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 10*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "taskmanager"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"),
			Port: getEnv("REDIS_PORT", "6379"),
			DB:   getIntEnv("REDIS_DB", 0),
		},
		Auth: AuthConfig{
			SigningKey: getEnv("JWT_SECRET", "your-secret-key"),
			TokenTTL:   getDurationEnv("JWT_EXPIRES", 24*time.Hour),
		},
		Logger: LoggerConfig{
			Level:       getEnv("LOG_LEVEL", "info"),
			File:        getEnv("LOG_FILE", ""),
			Format:      getEnv("LOG_FORMAT", "text"),
			ServiceName: getEnv("SERVICE_NAME", "task-manager"),
			Environment: getEnv("ENVIRONMENT", "development"),
		},
	}, nil
}

// ConnectionString возвращает строку подключения к PostgreSQL
func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getIntEnv возвращает значение переменной окружения как int
func getIntEnv(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// getDurationEnv возвращает значение переменной окружения как time.Duration
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
