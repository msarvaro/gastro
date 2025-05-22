package configs

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
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
		Port   string
		JWTKey string
	}
	Paths struct {
		ProjectRoot string
		Frontend    string
		Static      string
		Templates   string
	}
}

// LoadConfig loads configuration from .env file
func LoadConfig() (*Config, error) {
	config := &Config{}

	// Load .env file
	envFile := ".env"
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		envFile = "../.env" // Try one directory up if not found in current dir
		if _, err := os.Stat(envFile); os.IsNotExist(err) {
			envFile = "../../.env" // Try two directories up if not found in current dir
			if _, err := os.Stat(envFile); os.IsNotExist(err) {
				envFile = "../../../.env" // Try three directories up if not found in current dir
			}
		}
	}

	err := godotenv.Load(envFile)
	if err != nil {
		// In production environments like Render.com, .env file might not exist
		// Just log the error but continue with environment variables
		fmt.Printf("Warning: error loading .env file: %v\n", err)
	}

	// Database configuration
	var envErr error
	config.Database.Host, envErr = getRequiredEnv("DB_HOST")
	if envErr != nil {
		return nil, envErr
	}

	dbPortStr, envErr := getRequiredEnv("DB_PORT")
	if envErr != nil {
		return nil, envErr
	}
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT, must be an integer: %v", err)
	}
	config.Database.Port = dbPort

	config.Database.User, envErr = getRequiredEnv("DB_USER")
	if envErr != nil {
		return nil, envErr
	}

	config.Database.Password, envErr = getRequiredEnv("DB_PASSWORD")
	if envErr != nil {
		return nil, envErr
	}

	config.Database.DBName, envErr = getRequiredEnv("DB_NAME")
	if envErr != nil {
		return nil, envErr
	}

	config.Database.SSLMode, envErr = getRequiredEnv("DB_SSL_MODE")
	if envErr != nil {
		return nil, envErr
	}

	// Server configuration
	// Try to get PORT from environment (Render.com sets this)
	renderPort := os.Getenv("PORT")
	if renderPort != "" {
		config.Server.Port = renderPort
	} else {
		// Fall back to SERVER_PORT if PORT is not set
		config.Server.Port, envErr = getRequiredEnv("SERVER_PORT")
		if envErr != nil {
			return nil, envErr
		}
	}

	config.Server.JWTKey, envErr = getRequiredEnv("JWT_KEY")
	if envErr != nil {
		return nil, envErr
	}

	// Paths configuration
	projectRootPath, envErr := getRequiredEnv("PROJECT_ROOT")
	if envErr != nil {
		return nil, envErr
	}
	projectRoot, err := filepath.Abs(projectRootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get project root: %v", err)
	}

	frontendPath, envErr := getRequiredEnv("FRONTEND_PATH")
	if envErr != nil {
		return nil, envErr
	}

	config.Paths.ProjectRoot = projectRoot
	config.Paths.Frontend = filepath.Join(projectRoot, frontendPath)
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

// getRequiredEnv gets an environment variable or returns an error if it's not set
func getRequiredEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("required environment variable '%s' is not set", key)
	}
	return value, nil
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
