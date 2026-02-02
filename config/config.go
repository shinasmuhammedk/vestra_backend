package config

import (
	"os"

	"vestra-ecommerce/utils/logging"

	"gopkg.in/yaml.v3"
)

// ServerConfig holds server configuration
type ServerConfig struct {
	Port    int    `yaml:"port"`
	Prefork bool   `yaml:"prefork"`
}

// DBConfig holds database configuration
type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"sslmode"`
	TimeZone string `yaml:"timezone"`
}

// SMTPConfig holds email configuration
type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	AccessSecret     string `yaml:"access_secret"`
	RefreshSecret    string `yaml:"refresh_secret"`
	AccessTTLMinutes int    `yaml:"access_ttl_minutes"`
	RefreshTTLHours  int    `yaml:"refresh_ttl_hours"`
}

// Config holds all application configuration
type Config struct {
	Server ServerConfig `yaml:"server"`
	DB     DBConfig     `yaml:"db"`
	SMTP   SMTPConfig   `yaml:"smtp"`
	JWT    JWTConfig    `yaml:"jwt"`
}

// LoadConfig loads configuration from YAML file
func LoadConfig(path string) (*Config, error) {
	logging.Debug.Printf("Loading configuration from: %s", path)

	cfg := &Config{}

	file, err := os.ReadFile(path)
	if err != nil {
		logging.Error.Printf("Failed to read config file: %v", err)
		return nil, err
	}
	logging.Debug.Printf("Config file read successfully (%d bytes)", len(file))

	if err := yaml.Unmarshal(file, cfg); err != nil {
		logging.Error.Printf("Failed to unmarshal YAML config: %v", err)
		return nil, err
	}

	// Log sensitive information safely (mask passwords)
	logging.Debug.Printf("Server config: Port=%d, Prefork=%v", 
		cfg.Server.Port, cfg.Server.Prefork)
	logging.Debug.Printf("Database config: Host=%s, Port=%d, Name=%s, User=%s", 
		cfg.DB.Host, cfg.DB.Port, cfg.DB.Name, cfg.DB.User)
	logging.Debug.Printf("SMTP config: Host=%s, Port=%d, From=%s", 
		cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.From)
	logging.Debug.Printf("JWT config: AccessTTL=%d mins, RefreshTTL=%d hours", 
		cfg.JWT.AccessTTLMinutes, cfg.JWT.RefreshTTLHours)

	logging.Debug.Println("Configuration loaded successfully")
	return cfg, nil
}