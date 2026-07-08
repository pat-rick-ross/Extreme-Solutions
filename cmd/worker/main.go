package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"Extreme-Solutions/internal/config"
	"Extreme-Solutions/internal/pkg/logger"
	"Extreme-Solutions/internal/repository/postgres"
	"Extreme-Solutions/internal/worker"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
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

	// Initialize Asynq server
	asynqRedis := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	asynqServer := asynq.NewServer(
		asynqRedis,
		asynq.Config{
			Concurrency: cfg.Queue.Concurrency,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
				return time.Duration(n) * cfg.Queue.RetryDelay
			},
			MaxRetry: cfg.Queue.RetryMax,
		},
	)

	// Initialize repositories
	customerRepo := postgres.NewCustomerRepository(dbPool)
	invoiceRepo := postgres.NewInvoiceRepository(dbPool)
	paymentRepo := postgres.NewPaymentRepository(dbPool)
	packageRepo := postgres.NewPackageRepository(dbPool)

	// Initialize workers
	invoiceWorker := worker.NewInvoiceGenerator(customerRepo, invoiceRepo, paymentRepo)
	suspensionWorker := worker.NewSuspensionChecker(customerRepo, invoiceRepo)
	reconcilerWorker := worker.NewPaymentReconciler(paymentRepo, invoiceRepo, customerRepo)

	// Create mux and register handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(worker.TypeInvoiceGeneration, invoiceWorker.HandleInvoiceGeneration)
	mux.HandleFunc(worker.TypeSuspensionCheck, suspensionWorker.HandleSuspensionCheck)
	mux.HandleFunc(worker.TypePaymentReconciliation, reconcilerWorker.HandlePaymentReconciliation)

	// Start worker
	go func() {
		logger.Info("Starting worker...", nil)
		if err := asynqServer.Run(mux); err != nil {
			logger.Fatal("Worker failed", map[string]interface{}{"error": err})
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down worker...", nil)
	asynqServer.Shutdown()
	logger.Info("Worker exited properly", nil)
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
