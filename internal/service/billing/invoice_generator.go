package billing

import (
	"context"
	"fmt"
	"log"
	"time"

	"Extreme-Solutions/internal/domain"
	"Extreme-Solutions/internal/repository"
)

type InvoiceGenerator struct {
	invoiceRepo  repository.InvoiceRepository
	customerRepo repository.CustomerRepository
	packageRepo  repository.PackageRepository
}

func NewInvoiceGenerator(
	ir repository.InvoiceRepository,
	cr repository.CustomerRepository,
	pr repository.PackageRepository,
) *InvoiceGenerator {
	return &InvoiceGenerator{
		invoiceRepo:  ir,
		customerRepo: cr,
		packageRepo:  pr,
	}
}

// GenerateMonthlyInvoices scans for active clients and generates monthly billing logs
func (ig *InvoiceGenerator) GenerateMonthlyInvoices(ctx context.Context) error {
	// 1. Fetch all active customers from database storage repositories
	customers, err := ig.customerRepo.ListActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to pull active billing pool: %w", err)
	}

	now := time.Now()
	_, month, year := now.Date()

	for _, customer := range customers {
		// Safety Check: Verify the pointer is not nil before dereferencing
		if customer.PackageID == nil {
			log.Printf("[WARN] Skipping billing for customer %s; no package ID assigned", customer.ID.String())
			continue
		}

		// 2. Fetch package cost parameters by dereferencing the pointer safely using *
		pkg, err := ig.packageRepo.GetByID(ctx, *customer.PackageID)
		if err != nil {
			log.Printf("[ERROR] Skipping billing for customer %s; package resolution error: %v", customer.ID.String(), err)
			continue
		}

		// 3. Build invoice identification scheme (e.g., INV-2026-JUL-UUID)
		invoiceRef := fmt.Sprintf("INV-%d-%s-%s", year, month.String()[:3], customer.ID.String())

		// Check if invoice has already been provisioned to avoid double billing records
		exists, _ := ig.invoiceRepo.ExistsByReference(ctx, invoiceRef)
		if exists {
			continue
		}

		// Calculate payment deadlines (standard 5-day grace periods)
		dueDate := time.Date(year, month, 5, 23, 59, 59, 0, time.Local)

		inv := &domain.Invoice{
			CustomerID:  customer.ID,
			Reference:   invoiceRef,
			Amount:      pkg.Price,
			Status:      "pending",
			DueDate:     dueDate,
			PeriodStart: time.Date(year, month, 1, 0, 0, 0, 0, time.Local),
			PeriodEnd:   time.Date(year, month, 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, -1),
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := ig.invoiceRepo.Create(ctx, inv); err != nil {
			log.Printf("[CRITICAL] Could not persist invoice statement for user %s: %v", customer.ID.String(), err)
		}
	}

	return nil
}
