package configs

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Database struct {
		Host     string
		Port     int
		User     string
		Password string
		DBName   string
		SSLMode  string
	}
	Server struct {
		Port   int
		JWTKey string
	}
	Paths struct {
		ProjectRoot string
		Frontend    string
		Static      string
		Templates   string
	}
}

func LoadConfig() (*Config, error) {
	config := &Config{}

	// Database configuration
	config.Database.Host = "localhost"
	config.Database.Port = 5432
	config.Database.User = "postgres"
	config.Database.Password = "postgres"
	config.Database.DBName = "user_management"
	config.Database.SSLMode = "disable"

	// Server configuration
	config.Server.Port = 8080
	config.Server.JWTKey = "your-secret-key" // В продакшене использовать безопасный ключ

	// Paths configuration
	projectRoot, err := filepath.Abs("../../")
	if err != nil {
		return nil, fmt.Errorf("failed to get project root: %v", err)
	}

	config.Paths.ProjectRoot = projectRoot
	config.Paths.Frontend = filepath.Join(projectRoot, "frontend")
	config.Paths.Static = filepath.Join(config.Paths.Frontend, "static")
	config.Paths.Templates = filepath.Join(config.Paths.Frontend, "templates")

	// Проверка существования директорий
	dirs := []string{
		config.Paths.Frontend,
		config.Paths.Static,
		config.Paths.Templates,
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return nil, fmt.Errorf("directory not found: %s", dir)
		}
	}

	return config, nil
}

func (c *Config) GetDBConnString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}
