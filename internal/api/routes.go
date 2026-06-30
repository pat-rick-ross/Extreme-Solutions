package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/your-org/isp-billing/internal/api/handlers"
	"github.com/your-org/isp-billing/internal/api/middleware"
	"github.com/your-org/isp-billing/internal/config"
	"github.com/your-org/isp-billing/internal/repository"
	"github.com/your-org/isp-billing/internal/repository/redis"
	"github.com/your-org/isp-billing/internal/service/network"
	"github.com/your-org/isp-billing/internal/service/payment"
	"github.com/your-org/isp-billing/internal/service/billing"
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
	cache *redis.Cache,
	provisioner *network.Provisioner,
	mpesaService *payment.MPESAService,
	invoiceGenerator *billing.InvoiceGenerator,
	proRater *billing.ProRater,
) *Server {
	s := &Server{
		router: chi.NewRouter(),
		cfg:    cfg,
	}

	s.setupMiddleware()
	s.setupRoutes(customerRepo, invoiceRepo, paymentRepo, packageRepo, cache, provisioner, mpesaService, invoiceGenerator, proRater)

	return s
}

func (s *Server) setupMiddleware() {
	// CORS
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Standard middleware
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
	cache *redis.Cache,
	provisioner *network.Provisioner,
	mpesaService *payment.MPESAService,
	invoiceGenerator *billing.InvoiceGenerator,
	proRater *billing.ProRater,
) {
	// Initialize handlers
	customerHandler := handlers.NewCustomerHandler(customerRepo, packageRepo)
	invoiceHandler := handlers.NewInvoiceHandler(invoiceRepo, customerRepo)
	paymentHandler := handlers.NewPaymentHandler(paymentRepo, invoiceRepo, customerRepo, mpesaService)
	authHandler := handlers.NewAuthHandler(customerRepo, cache)
	webhookHandler := handlers.NewWebhookHandler(paymentRepo, invoiceRepo, customerRepo)

	// Rate limiter
	rateLimiter := middleware.NewRateLimiter(redisClient, 100, time.Minute)

	// Public routes
	s.router.Group(func(r chi.Router) {
		r.Post("/api/v1/auth/login", authHandler.Login)
		r.Post("/api/v1/auth/register", authHandler.Register)
		r.Post("/api/v1/auth/refresh", authHandler.Refresh)
	})

	// Webhook routes (no auth)
	s.router.Group(func(r chi.Router) {
		r.Post("/webhook/mpesa/validation", webhookHandler.Validate)
		r.Post("/webhook/mpesa/confirmation", webhookHandler.Confirm)
		r.Post("/webhook/mpesa/callback", webhookHandler.Callback)
		r.Post("/webhook/mpesa/timeout", webhookHandler.Timeout)
	})

	// Protected routes
	s.router.Group(func(r chi.Router) {
		r.Use(middleware.Auth)
		r.Use(rateLimiter.Limit)

		// Customer routes
		r.Route("/api/v1/customers", func(r chi.Router) {
			r.Get("/", customerHandler.List)
			r.Post("/", customerHandler.Create)
			r.Get("/{id}", customerHandler.GetByID)
			r.Put("/{id}", customerHandler.Update)
			r.Delete("/{id}", customerHandler.Delete)
		})

		// Package routes
		r.Route("/api/v1/packages", func(r chi.Router) {
			r.Get("/", packageHandler.List)
			r.Post("/", packageHandler.Create)
			r.Get("/{id}", packageHandler.GetByID)
			r.Put("/{id}", packageHandler.Update)
			r.Delete("/{id}", packageHandler.Delete)
		})

		// Invoice routes
		r.Route("/api/v1/invoices", func(r chi.Router) {
			r.Get("/", invoiceHandler.List)
			r.Get("/{id}", invoiceHandler.GetByID)
			r.Get("/{id}/pdf", invoiceHandler.GetPDF)
			r.Post("/generate", invoiceHandler.Generate)
		})

		// Payment routes
		r.Route("/api/v1/payments", func(r chi.Router) {
			r.Post("/initiate", paymentHandler.InitiatePayment)
			r.Get("/{id}", paymentHandler.GetPaymentStatus)
			r.Get("/customer", paymentHandler.GetCustomerPayments)
		})

		// Network routes
		r.Route("/api/v1/network", func(r chi.Router) {
			r.Post("/provision", networkHandler.Provision)
			r.Post("/suspend/{customer_id}", networkHandler.Suspend)
			r.Post("/reactivate/{customer_id}", networkHandler.Reactivate)
		})
	})

	// Health check
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	s.router.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		// Check database connectivity
		if err := dbPool.Ping(context.Background()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"not ready","error":"database unavailable"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})
}

func (s *Server) Routes() http.Handler {
	return s.router
}
