package api

import (
	"net/http"
	"time"

	"Extreme-Solutions/internal/api/handlers"
	customMiddleware "Extreme-Solutions/internal/api/middleware"
	"Extreme-Solutions/internal/config"
	"Extreme-Solutions/internal/repository"
	"Extreme-Solutions/internal/service/billing"
	"Extreme-Solutions/internal/service/network"
	"Extreme-Solutions/internal/service/payment"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
)

type Server struct {
	router *chi.Mux
	cfg    *config.Config
}

func NewServer(
	cfg *config.Config,
	customerRepo repository.CustomerRepository,
	invoiceRepo repository.InvoiceRepository,
	paymentRepo repository.PaymentRepository,
	packageRepo repository.PackageRepository,
	cache repository.CacheRepository,
	redisClient *redis.Client, // Concrete type injection received here
	provisioner *network.Provisioner,
	darajaSvc *payment.DarajaService,
	paystackSvc *payment.PaystackService,
	reconciler *payment.PaymentReconciler,
	invoiceGenerator *billing.InvoiceGenerator,
	proRater *billing.ProRater,
) *Server {
	s := &Server{
		router: chi.NewRouter(),
		cfg:    cfg,
	}

	s.setupMiddleware()
	// Fixed: Added redisClient to the setupRoutes call parameter sequence
	s.setupRoutes(customerRepo, invoiceRepo, paymentRepo, packageRepo, cache, redisClient, provisioner, darajaSvc, paystackSvc, reconciler, invoiceGenerator, proRater)

	return s
}

func (s *Server) setupMiddleware() {
	s.router.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "x-paystack-signature"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}).Handler)

	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))
}

func (s *Server) setupRoutes(
	customerRepo repository.CustomerRepository,
	invoiceRepo repository.InvoiceRepository,
	paymentRepo repository.PaymentRepository,
	packageRepo repository.PackageRepository,
	cache repository.CacheRepository,
	redisClient *redis.Client, // Fixed: Added signature definition here to wire up line 90
	provisioner *network.Provisioner,
	darajaSvc *payment.DarajaService,
	paystackSvc *payment.PaystackService,
	reconciler *payment.PaymentReconciler,
	invoiceGenerator *billing.InvoiceGenerator,
	proRater *billing.ProRater,
) {
	// Initialize domain feature handlers
	customerHandler := handlers.NewCustomerHandler(customerRepo, packageRepo)
	invoiceHandler := handlers.NewInvoiceHandler(invoiceRepo, customerRepo)
	paymentHandler := handlers.NewPaymentHandler(paymentRepo, invoiceRepo, customerRepo, darajaSvc, paystackSvc)
	authHandler := handlers.NewAuthHandler(customerRepo, cache)
	webhookHandler := handlers.NewWebhookHandler(paystackSvc, darajaSvc, reconciler)

	// App Rate limiter mapping (Now cleanly picks up the injected parameter context)
	rateLimiter := customMiddleware.NewRateLimiter(redisClient, 100, time.Minute)

	// Public routes
	s.router.Group(func(r chi.Router) {
		r.Post("/api/v1/auth/login", authHandler.Login)
		r.Post("/api/v1/auth/register", authHandler.Register)
		r.Post("/api/v1/auth/refresh", authHandler.Refresh)
	})

	// Webhook routes (Completely public with internal cryptographic validation)
	s.router.Group(func(r chi.Router) {
		r.Post("/api/v1/payments/webhook/daraja", webhookHandler.HandleDarajaWebhook)
		r.Post("/api/v1/payments/webhook/paystack", webhookHandler.HandlePaystackWebhook)
	})

	// Protected routes
	s.router.Group(func(r chi.Router) {
		r.Use(customMiddleware.Auth(s.cfg.JWT.Secret))
		r.Use(rateLimiter.Limit)

		// Customer routing definitions
		r.Route("/api/v1/customers", func(r chi.Router) {
			r.Get("/", customerHandler.List)
			r.Post("/", customerHandler.Create)
			r.Get("/{id}", customerHandler.GetByID)
			r.Put("/{id}", customerHandler.Update)
			r.Delete("/{id}", customerHandler.Delete)
		})

		// Invoice execution routes
		r.Route("/api/v1/invoices", func(r chi.Router) {
			r.Get("/", invoiceHandler.List)
			r.Get("/{id}", invoiceHandler.GetByID)
			r.Get("/{id}/pdf", invoiceHandler.GetPDF)
			r.Post("/generate", invoiceHandler.Generate)
		})

		// Multi-Gateway explicit payment invocation paths
		r.Route("/api/v1/payments", func(r chi.Router) {
			r.Post("/initiate", paymentHandler.InitiatePayment)
			r.Get("/{id}", paymentHandler.GetPaymentStatus)
			r.Get("/customer", paymentHandler.GetCustomerPayments)
		})
	})

	// Microservices health indicators
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
}

func (s *Server) Routes() http.Handler {
	return s.router
}
