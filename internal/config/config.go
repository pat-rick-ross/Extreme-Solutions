package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Mikrotik MikrotikConfig `yaml:"mikrotik"`
	Mpesa    MpesaConfig    `yaml:"mpesa"`
	SMS      SMSConfig      `yaml:"sms"`
	Email    EmailConfig    `yaml:"email"`
	Queue    QueueConfig    `yaml:"queue"`
	Paystack PaystackConfig `yaml:"paystack"` // <-- MUST BE CAPITALIZED 'Paystack' here
	IntaSend IntaSendConfig `yaml:"intasend"` // Add this
}

type AppConfig struct {
	Env         string `yaml:"env"`
	Port        int    `yaml:"port"`
	LogLevel    string `yaml:"log_level"`
	APIBaseURL  string `yaml:"api_base_url"`
	FrontendURL string `yaml:"frontend_url"`
}

type DatabaseConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	User         string        `yaml:"user"`
	Password     string        `yaml:"password"`
	DBName       string        `yaml:"dbname"`
	SSLMode      string        `yaml:"sslmode"`
	MaxOpenConns int           `yaml:"max_open_conns"`
	MaxIdleConns int           `yaml:"max_idle_conns"`
	MaxLifetime  time.Duration `yaml:"max_lifetime"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type JWTConfig struct {
	Secret          string        `yaml:"secret"`
	AccessDuration  time.Duration `yaml:"access_duration"`
	RefreshDuration time.Duration `yaml:"refresh_duration"`
}

type MikrotikConfig struct {
	Host     string        `yaml:"host"`
	Port     int           `yaml:"port"`
	Username string        `yaml:"username"`
	Password string        `yaml:"password"`
	Timeout  time.Duration `yaml:"timeout"`
}

type MpesaConfig struct {
	ConsumerKey     string `yaml:"consumer_key"`
	ConsumerSecret  string `yaml:"consumer_secret"`
	Passkey         string `yaml:"passkey"`
	Shortcode       string `yaml:"shortcode"`
	Environment     string `yaml:"environment"`
	CallbackURL     string `yaml:"callback_url"`
	TimeoutURL      string `yaml:"timeout_url"`
	ResultURL       string `yaml:"result_url"`
	ConfirmationURL string `yaml:"confirmation_url"`
	ValidationURL   string `yaml:"validation_url"`
}

type PaystackConfig struct {
	SecretKey  string `yaml:"secret_key"`
	PublicKey  string `yaml:"public_key"`
	WebhookURL string `yaml:"webhook_url"`
}

type IntaSendConfig struct {
	PublishableKey string `yaml:"publishable_key"`
	SecretKey      string `yaml:"secret_key"`
	Environment    string `yaml:"environment"`
}

type SMSConfig struct {
	Provider   string `yaml:"provider"`
	APIKey     string `yaml:"api_key"`
	SenderID   string `yaml:"sender_id"`
	Enabled    bool   `yaml:"enabled"`
	RetryCount int    `yaml:"retry_count"`
}

type EmailConfig struct {
	SMTPHost string `yaml:"smtp_host"`
	SMTPPort int    `yaml:"smtp_port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
	Enabled  bool   `yaml:"enabled"`
}

type QueueConfig struct {
	Concurrency int           `yaml:"concurrency"`
	RetryMax    int           `yaml:"retry_max"`
	RetryDelay  time.Duration `yaml:"retry_delay"`
}

// Load reads config from a YAML file path
func Load(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config YAML: %w", err)
	}

	return &cfg, nil
}
