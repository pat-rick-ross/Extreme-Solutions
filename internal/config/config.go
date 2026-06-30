package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	MikroTik MikroTikConfig
	MPESA    MPESAConfig
	SMS      SMSConfig
	Email    EmailConfig
	Queue    QueueConfig
}

type AppConfig struct {
	Env          string
	Port         int
	LogLevel     string
	APIBaseURL   string
	FrontendURL  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type JWTConfig struct {
	Secret          string
	AccessDuration  time.Duration
	RefreshDuration time.Duration
}

type MikroTikConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Timeout  time.Duration
}

type MPESAConfig struct {
	ConsumerKey     string
	ConsumerSecret  string
	Passkey         string
	ShortCode       string
	Environment     string
	CallbackURL     string
	TimeoutURL      string
	ResultURL       string
	ConfirmationURL string
	ValidationURL   string
}

type SMSConfig struct {
	Provider   string
	APIKey     string
	SenderID   string
	Enabled    bool
	RetryCount int
}

type EmailConfig struct {
	SMTPHost string
	SMTPPort int
	Username string
	Password string
	From     string
	Enabled  bool
}

type QueueConfig struct {
	Concurrency int
	RetryMax    int
	RetryDelay  time.Duration
}

var config *Config

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() *Config {
	cfg := &Config{}

	// App Config
	cfg.App.Env = getEnv("APP_ENV", "development")
	cfg.App.Port = getEnvAsInt("APP_PORT", 8080)
	cfg.App.LogLevel = getEnv("APP_LOG_LEVEL", "debug")
	cfg.App.APIBaseURL = getEnv("APP_API_BASE_URL", "http://localhost:8080")
	cfg.App.FrontendURL = getEnv("APP_FRONTEND_URL", "http://localhost:3000")
	cfg.App.ReadTimeout = getEnvAsDuration("HTTP_READ_TIMEOUT", 15*time.Second)
	cfg.App.WriteTimeout = getEnvAsDuration("HTTP_WRITE_TIMEOUT", 15*time.Second)
	cfg.App.IdleTimeout = getEnvAsDuration("HTTP_IDLE_TIMEOUT", 60*time.Second)

	// Database Config
	cfg.Database.Host = getEnv("DB_HOST", "localhost")
	cfg.Database.Port = getEnvAsInt("DB_PORT", 5432)
	cfg.Database.User = getEnv("DB_USER", "isp_user")
	cfg.Database.Password = getEnv("DB_PASSWORD", "")
	cfg.Database.DBName = getEnv("DB_NAME", "isp_billing")
	cfg.Database.SSLMode = getEnv("DB_SSLMODE", "disable")
	cfg.Database.MaxOpenConns = getEnvAsInt("DB_MAX_OPEN_CONNS", 25)
	cfg.Database.MaxIdleConns = getEnvAsInt("DB_MAX_IDLE_CONNS", 5)
	cfg.Database.MaxLifetime = getEnvAsDuration("DB_MAX_LIFETIME", 5*time.Minute)

	// Redis Config
	cfg.Redis.Host = getEnv("REDIS_HOST", "localhost")
	cfg.Redis.Port = getEnvAsInt("REDIS_PORT", 6379)
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.DB = getEnvAsInt("REDIS_DB", 0)

	// JWT Config
	cfg.JWT.Secret = getEnv("JWT_SECRET", "change-this-in-production")
	cfg.JWT.AccessDuration = getEnvAsDuration("JWT_ACCESS_DURATION", 15*time.Minute)
	cfg.JWT.RefreshDuration = getEnvAsDuration("JWT_REFRESH_DURATION", 168*time.Hour)

	// MikroTik Config
	cfg.MikroTik.Host = getEnv("MIKROTIK_HOST", "192.168.88.1")
	cfg.MikroTik.Port = getEnvAsInt("MIKROTIK_PORT", 8728)
	cfg.MikroTik.Username = getEnv("MIKROTIK_USERNAME", "admin")
	cfg.MikroTik.Password = getEnv("MIKROTIK_PASSWORD", "")
	cfg.MikroTik.Timeout = getEnvAsDuration("MIKROTIK_TIMEOUT", 30*time.Second)

	// M-PESA Config (Updated with your sandbox credentials)
	cfg.MPESA.ConsumerKey = getEnv("MPESA_CONSUMER_KEY", "mFPUNIJtNTufp6fTmDGtLqUNsAFKhoicg")
	cfg.MPESA.ConsumerSecret = getEnv("MPESA_CONSUMER_SECRET", "3JPrzACzDKJ2ZDOLXk5zLtfJGXP9yEbv")
	cfg.MPESA.Passkey = getEnv("MPESA_PASSKEY", "")
	cfg.MPESA.ShortCode = getEnv("MPESA_SHORTCODE", "174379")
	cfg.MPESA.Environment = getEnv("MPESA_ENVIRONMENT", "sandbox")
	cfg.MPESA.CallbackURL = getEnv("MPESA_CALLBACK_URL", "https://your-domain.com/webhook/mpesa/callback")
	cfg.MPESA.TimeoutURL = getEnv("MPESA_TIMEOUT_URL", "https://your-domain.com/webhook/mpesa/timeout")
	cfg.MPESA.ResultURL = getEnv("MPESA_RESULT_URL", "https://your-domain.com/webhook/mpesa/result")
	cfg.MPESA.ConfirmationURL = getEnv("MPESA_CONFIRMATION_URL", "https://your-domain.com/webhook/mpesa/confirmation")
	cfg.MPESA.ValidationURL = getEnv("MPESA_VALIDATION_URL", "https://your-domain.com/webhook/mpesa/validation")

	// SMS Config
	cfg.SMS.Provider = getEnv("SMS_PROVIDER", "africastalking")
	cfg.SMS.APIKey = getEnv("SMS_API_KEY", "")
	cfg.SMS.SenderID = getEnv("SMS_SENDER_ID", "YOUR_ISP")
	cfg.SMS.Enabled = getEnvAsBool("SMS_ENABLED", true)
	cfg.SMS.RetryCount = getEnvAsInt("SMS_RETRY_COUNT", 3)

	// Email Config
	cfg.Email.SMTPHost = getEnv("EMAIL_SMTP_HOST", "smtp.gmail.com")
	cfg.Email.SMTPPort = getEnvAsInt("EMAIL_SMTP_PORT", 587)
	cfg.Email.Username = getEnv("EMAIL_USERNAME", "")
	cfg.Email.Password = getEnv("EMAIL_PASSWORD", "")
	cfg.Email.From = getEnv("EMAIL_FROM", "support@yourisp.com")
	cfg.Email.Enabled = getEnvAsBool("EMAIL_ENABLED", true)

	// Queue Config
	cfg.Queue.Concurrency = getEnvAsInt("QUEUE_CONCURRENCY", 10)
	cfg.Queue.RetryMax = getEnvAsInt("QUEUE_RETRY_MAX", 3)
	cfg.Queue.RetryDelay = getEnvAsDuration("QUEUE_RETRY_DELAY", 5*time.Minute)

	// Validate required fields
	cfg.validate()

	config = cfg
	return cfg
}

func (c *Config) validate() {
	if c.Database.Password == "" {
		log.Println("WARNING: DB_PASSWORD is not set!")
	}

	if c.JWT.Secret == "change-this-in-production" && c.App.Env == "production" {
		log.Println("WARNING: JWT_SECRET is using default value in production!")
	}

	if c.MPESA.ConsumerKey == "" || c.MPESA.ConsumerSecret == "" {
		log.Println("WARNING: M-PESA credentials are not configured!")
	}

	if c.MPESA.Passkey == "" {
		log.Println("WARNING: M-PESA Passkey is not set! You need to set this for STK Push to work.")
	}
}

func Get() *Config {
	if config == nil {
		LoadFromEnv()
	}
	return config
}

// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}
