package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/your-org/isp-billing/internal/api"
	"github.com/your-org/isp-billing/internal/config"
	"github.com/your-org/isp-billing/internal/pkg/logger"
	"github.com/your-org/isp-billing/internal/repository/postgres"
	redisRepo "github.com/your-org/isp-billing/internal/repository/redis"
	"github.com/your-org/isp-billing/internal/service/billing"
	"github.com/your-org/isp-billing/internal/service/network"
	"github.com/your-org/isp-billing/internal/service/payment"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Load configuration from env
	cfg := config.LoadFromEnv()

	// Initialize logger
	logger.Init(cfg.App.LogLevel)

	// Connect to PostgreSQL
	dbPool, err := pgxpool.New(context.Background(), getDBConnectionString(cfg))
	if err != nil {
		logger.Fatal("Failed to connect to database", map[string]interface{}{"error": err})
	}
	defer dbPool.Close()

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// Initialize repositories
	customerRepo := postgres.NewCustomerRepository(dbPool)
	invoiceRepo := postgres.NewInvoiceRepository(dbPool)
	paymentRepo := postgres.NewPaymentRepository(dbPool)
	packageRepo := postgres.NewPackageRepository(dbPool)

	// Initialize cache
	cache := redisRepo.NewCache(redisClient)

	// Initialize services
	provisioner := network.NewProvisioner(cfg.MikroTik)
	mpesaService := payment.NewMPESAService(cfg.MPESA)
	invoiceGenerator := billing.NewInvoiceGenerator(invoiceRepo, customerRepo, paymentRepo, cache)
	proRater := billing.NewProRater()

	// Initialize API server
	server := api.NewServer(
		cfg,
		customerRepo,
		invoiceRepo,
		paymentRepo,
		packageRepo,
		cache,
		provisioner,
		mpesaService,
		invoiceGenerator,
		proRater,
	)

	// Start HTTP server
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      server.Routes(),
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
		IdleTimeout:  cfg.App.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting API server", map[string]interface{}{"port": cfg.App.Port})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", map[string]interface{}{"error": err})
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced shutdown", map[string]interface{}{"error": err})
	}

	logger.Info("Server exited properly", nil)
}

func getDBConnectionString(cfg *config.Config) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&pool_max_conns=%d&pool_max_conn_idle_time=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
		cfg.Database.MaxOpenConns,
		cfg.Database.MaxLifetime.String(),
	)
}
