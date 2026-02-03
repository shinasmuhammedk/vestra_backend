package config

import (
	"os"

	"vestra-ecommerce/helper"
	"vestra-ecommerce/utils/logging"

	"gopkg.in/yaml.v3"
)

/*
|--------------------------------------------------------------------------
| Server Configuration
|--------------------------------------------------------------------------
*/
type ServerConfig struct {
	Port    int  `yaml:"port"`    // Server port (e.g. 8080)
	Prefork bool `yaml:"prefork"` // Enable prefork (Fiber optimization)
}

/*
|--------------------------------------------------------------------------
| Database Configuration
|--------------------------------------------------------------------------
*/
type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"sslmode"`
	TimeZone string `yaml:"timezone"`
}

/*
|--------------------------------------------------------------------------
| SMTP / Email Configuration
|--------------------------------------------------------------------------
*/
type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

/*
|--------------------------------------------------------------------------
| JWT Configuration
|--------------------------------------------------------------------------
*/
type JWTConfig struct {
	AccessSecret     string `yaml:"access_secret"`
	RefreshSecret    string `yaml:"refresh_secret"`
	AccessTTLMinutes int    `yaml:"access_ttl_minutes"`
	RefreshTTLHours  int    `yaml:"refresh_ttl_hours"`
}

/*
|--------------------------------------------------------------------------
| Razorpay Configuration (TEST / LIVE)
|--------------------------------------------------------------------------
| Use Razorpay TEST keys in development
| Switch to LIVE keys in production
*/
type RazorpayConfig struct {
	KeyID     string `yaml:"key_id"`
	KeySecret string `yaml:"key_secret"`
}


// Root Application Config

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	DB       DBConfig       `yaml:"db"`
	SMTP     SMTPConfig     `yaml:"smtp"`
	JWT      JWTConfig      `yaml:"jwt"`
	Razorpay RazorpayConfig `yaml:"razorpay"`
}

/*
 LoadConfig
 Reads YAML config file and unmarshals into Config struct
*/
func LoadConfig(path string) (*Config, error) {
	logging.Debug.Printf("📦 Loading configuration from: %s", path)

	cfg := &Config{}

	// Read file
	file, err := os.ReadFile(path)
	if err != nil {
		logging.Error.Printf("❌ Failed to read config file: %v", err)
		return nil, err
	}
	logging.Debug.Printf("✅ Config file read successfully (%d bytes)", len(file))

	// Unmarshal YAML
	if err := yaml.Unmarshal(file, cfg); err != nil {
		logging.Error.Printf("❌ Failed to unmarshal YAML config: %v", err)
		return nil, err
	}

	// Safe logging (never log secrets)
	logging.Debug.Printf(
		"🖥 Server: Port=%d | Prefork=%v",
		cfg.Server.Port,
		cfg.Server.Prefork,
	)

	logging.Debug.Printf(
		"🗄 Database: Host=%s | Port=%d | Name=%s | User=%s",
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
		cfg.DB.User,
	)

	logging.Debug.Printf(
		"📧 SMTP: Host=%s | Port=%d | From=%s",
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.From,
	)

	logging.Debug.Printf(
		"🔐 JWT: AccessTTL=%d mins | RefreshTTL=%d hours",
		cfg.JWT.AccessTTLMinutes,
		cfg.JWT.RefreshTTLHours,
	)

	logging.Debug.Printf(
		"💳 Razorpay: KeyID=%s",
		helper.MaskSecret(cfg.Razorpay.KeyID),
	)

	logging.Debug.Println("✅ Configuration loaded successfully")

	return cfg, nil
}
