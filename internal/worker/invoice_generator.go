package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"Extreme-Solutions/internal/domain"
	"Extreme-Solutions/internal/repository"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const TypeInvoiceGeneration = "invoice:generate"

type InvoiceGenerationPayload struct {
	CustomerID string `json:"customer_id"`
	Month      int    `json:"month"`
	Year       int    `json:"year"`
}

type InvoiceGenerator struct {
	customerRepo repository.CustomerRepository
	invoiceRepo  repository.InvoiceRepository
	packageRepo  repository.PackageRepository // Added package visibility
}

func NewInvoiceGenerator(
	customerRepo repository.CustomerRepository,
	invoiceRepo repository.InvoiceRepository,
	packageRepo repository.PackageRepository,
) *InvoiceGenerator {
	return &InvoiceGenerator{
		customerRepo: customerRepo,
		invoiceRepo:  invoiceRepo,
		packageRepo:  packageRepo,
	}
}

func (w *InvoiceGenerator) HandleInvoiceGeneration(ctx context.Context, t *asynq.Task) error {
	var payload InvoiceGenerationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("[INFO] Generating invoice for Customer ID: %s, Period: %02d/%d", payload.CustomerID, payload.Month, payload.Year)

	customerID, err := uuid.Parse(payload.CustomerID)
	if err != nil {
		return fmt.Errorf("invalid customer ID: %w", err)
	}

	customer, err := w.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}
	if customer == nil {
		return fmt.Errorf("customer not found")
	}

	// 1. Safe extraction of price metrics from related subscription package
	var packagePrice float64 = 0.0
	packageName := "Base Plan"

	if customer.PackageID != nil {
		pkg, err := w.packageRepo.GetByID(ctx, *customer.PackageID)
		if err != nil {
			return fmt.Errorf("failed to load package details: %w", err)
		}
		if pkg != nil {
			packagePrice = pkg.Price
			packageName = pkg.Name
		}
	}

	invoiceNumber := generateInvoiceNumber()

	// 2. Map directly to fields visible in image_0df83a.png
	invoice := &domain.Invoice{
		ID:          uuid.New(),
		CustomerID:  customer.ID,
		Number:      invoiceNumber,
		Amount:      packagePrice,
		Tax:         packagePrice * 0.16, // 16% VAT
		Total:       packagePrice * 1.16,
		Status:      "pending", // Inline string literal matching comment in image
		Description: fmt.Sprintf("ISP Service - %s (%s)", packageName, time.Now().Format("January 2006")),
		PeriodStart: time.Date(payload.Year, time.Month(payload.Month), 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(payload.Year, time.Month(payload.Month+1), 0, 23, 59, 59, 0, time.UTC),
		DueDate:     time.Date(payload.Year, time.Month(payload.Month), 15, 0, 0, 0, 0, time.UTC),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := w.invoiceRepo.Create(ctx, invoice); err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	log.Printf("[INFO] Invoice %s generated successfully. Total: %.2f", invoice.Number, invoice.Total)
	return nil
}

func generateInvoiceNumber() string {
	return fmt.Sprintf("INV-%d%02d%02d-%04d",
		time.Now().Year(),
		time.Now().Month(),
		time.Now().Day(),
		time.Now().UnixNano()%10000,
	)
}
