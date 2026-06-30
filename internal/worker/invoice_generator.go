package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/your-org/isp-billing/internal/domain"
	"github.com/your-org/isp-billing/internal/pkg/logger"
	"github.com/your-org/isp-billing/internal/repository"
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
	paymentRepo  repository.PaymentRepository
}

func NewInvoiceGenerator(
	customerRepo repository.CustomerRepository,
	invoiceRepo repository.InvoiceRepository,
	paymentRepo repository.PaymentRepository,
) *InvoiceGenerator {
	return &InvoiceGenerator{
		customerRepo: customerRepo,
		invoiceRepo:  invoiceRepo,
		paymentRepo:  paymentRepo,
	}
}

func (w *InvoiceGenerator) HandleInvoiceGeneration(ctx context.Context, t *asynq.Task) error {
	var payload InvoiceGenerationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	logger.Info("Generating invoice", map[string]interface{}{
		"customer_id": payload.CustomerID,
		"month":       payload.Month,
		"year":        payload.Year,
	})

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

	// Check if invoice already exists for this period
	// This is simplified - you'd need to add a method to check by period

	// Generate invoice number
	invoiceNumber := generateInvoiceNumber()

	invoice := &domain.Invoice{
		CustomerID:  customer.ID,
		Number:      invoiceNumber,
		Amount:      customer.Package.Price,
		Tax:         customer.Package.Price * 0.16, // 16% VAT
		Total:       customer.Package.Price * 1.16,
		Description: fmt.Sprintf("ISP Service - %s (%s)", customer.Package.Name, time.Now().Format("January 2006")),
		PeriodStart: time.Date(payload.Year, time.Month(payload.Month), 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(payload.Year, time.Month(payload.Month+1), 0, 23, 59, 59, 0, time.UTC),
		DueDate:     time.Date(payload.Year, time.Month(payload.Month), 15, 0, 0, 0, 0, time.UTC),
		Status:      domain.InvoiceStatusPending,
	}

	if err := w.invoiceRepo.Create(ctx, invoice); err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	logger.Info("Invoice generated successfully", map[string]interface{}{
		"invoice_id": invoice.ID,
		"number":     invoice.Number,
		"amount":     invoice.Total,
	})

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
