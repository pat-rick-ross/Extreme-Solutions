package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// External packages
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	// Workspace package references
	"Extreme-Solutions/internal/api"
	"Extreme-Solutions/internal/config"
	"Extreme-Solutions/internal/repository/postgres"
	"Extreme-Solutions/internal/service/billing"
	"Extreme-Solutions/internal/service/network"
	"Extreme-Solutions/internal/service/payment"
	"Extreme-Solutions/internal/worker"
)

func main() {
	// 1. Load .env file explicitly into environment memory layer if you use one
	if err := godotenv.Load(); err != nil {
		log.Println("[INFO] No .env file found, relying purely on YAML and system environment variables")
	}

	// 2. Point directly to your active config.yaml layout path
	cfg, err := config.Load("internal/config/config.yaml") // Aligned path parameter
	if err != nil {
		log.Fatalf("Critical Error: Failed to parse configuration specifications from file: %v", err)
	}

	// 3. Connect to PostgreSQL via database/sql pgx wrapper
	connString := getDBConnectionString(cfg)
	db, err := sql.Open("pgx", connString)
	if err != nil {
		log.Fatalf("Failed to open connection database handle: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxOpenConns / 2)

	// 4. Connect to Redis Instance
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// ==========================================
	// LAYERS ARCHITECTURE INITIALIZATION
	// ==========================================

	customerRepo := postgres.NewCustomerRepository(db)
	invoiceRepo := postgres.NewInvoiceRepository(db)
	paymentRepo := postgres.NewPaymentRepository(db)
	packageRepo := postgres.NewPackageRepository(db)
	cache := postgres.NewCache(redisClient)

	provisioner := network.NewProvisioner(cfg.Mikrotik)

	// Initialize Payment Services
	darajaSvc := payment.NewDarajaService(cfg)
	paystackSvc := payment.NewPaystackService(cfg)
	intasendSvc := payment.NewIntaSendService(cfg)

	// FIX: Match the required 4-argument constructor signature
	reconciler := payment.NewPaymentReconciler(
		paymentRepo,
		invoiceRepo,
		customerRepo,
		provisioner,
	)

	proRater := billing.NewProRater()
	invoiceGenerator := billing.NewInvoiceGenerator(invoiceRepo, customerRepo, packageRepo)

	// ... rest of your code remains the same ...
	// Background Automated Suspension Loop
	suspensionWorker := worker.NewSuspensionChecker(invoiceRepo, customerRepo, provisioner, 1*time.Hour)
	suspensionWorker.Start(context.Background())

	// Initialize API router container instance
	// Update these parameters to match your NewServer signature
	server := api.NewServer(
		cfg,
		customerRepo,
		invoiceRepo,
		paymentRepo,
		packageRepo,
		cache,
		redisClient,
		provisioner,
		darajaSvc,
		paystackSvc, // Now passing initialized Paystack
		intasendSvc, // Now passing initialized IntaSend
		reconciler,
		invoiceGenerator,
		proRater,
	)

	// ==========================================
	// RUNTIME LIFECYCLE SERVING
	// ==========================================

	addr := fmt.Sprintf(":%d", cfg.App.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      server.Routes(),
		ReadTimeout:  15 * time.Second, // Fixed: Uses clean explicit standard fallbacks
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[INFO] Extreme Solutions Engine online, routing incoming requests at port %d", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP framework engine: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[INFO] Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced structural failure on shutdown: %v", err)
	}

	log.Println("[INFO] Server exited cleanly. All interface connections closed.")
}

func getDBConnectionString(cfg *config.Config) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)
}
