package worker

import (
	"context"
	"database/sql"
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

func (ig *InvoiceGenerator) GenerateMonthlyInvoices(ctx context.Context) error {
	customers, err := ig.customerRepo.ListActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to pull active billing pool: %w", err)
	}

	now := time.Now()
	_, month, year := now.Date()

	for _, customer := range customers {
		if customer.PackageID == nil {
			log.Printf("[WARN] Skipping billing for customer %s; no package ID assigned", customer.ID.String())
			continue
		}

		pkg, err := ig.packageRepo.GetByID(ctx, *customer.PackageID)
		if err != nil {
			log.Printf("[ERROR] Skipping billing for customer %s; package resolution error: %v", customer.ID.String(), err)
			continue
		}

		invoiceRef := fmt.Sprintf("INV-%d-%s-%s", year, month.String()[:3], customer.ID.String())

		exists, _ := ig.invoiceRepo.ExistsByReference(ctx, invoiceRef)
		if exists {
			continue
		}

		dueDate := time.Date(year, month, 5, 23, 59, 59, 0, time.Local)

		// Updated struct initialization with sql.NullString
		inv := &domain.Invoice{
			CustomerID:  customer.ID,
			Number:      invoiceRef,
			Amount:      pkg.Price,
			Total:       pkg.Price,
			Status:      "pending",
			DueDate:     dueDate,
			PeriodStart: time.Date(year, month, 1, 0, 0, 0, 0, time.Local),
			PeriodEnd:   time.Date(year, month, 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, -1),
			CreatedAt:   now,
			UpdatedAt:   now,
			Description: sql.NullString{
				String: fmt.Sprintf("ISP Service - %s (%s)", pkg.Name, month.String()),
				Valid:  true,
			},
		}

		if err := ig.invoiceRepo.Create(ctx, inv); err != nil {
			log.Printf("[CRITICAL] Could not persist invoice statement for user %s: %v", customer.ID.String(), err)
		}
	}

	return nil
}
