package worker

import (
	"context"
	"fmt"
	"time"

	"Extreme-Solutions/internal/repository"
	"Extreme-Solutions/internal/service/network"
)

type SuspensionChecker struct {
	invoiceRepo  repository.InvoiceRepository
	customerRepo repository.CustomerRepository
	provisioner  *network.Provisioner
	interval     time.Duration
}

func NewSuspensionChecker(
	ir repository.InvoiceRepository,
	cr repository.CustomerRepository,
	provisioner *network.Provisioner,
	interval time.Duration,
) *SuspensionChecker {
	return &SuspensionChecker{
		invoiceRepo:  ir,
		customerRepo: cr,
		provisioner:  provisioner,
		interval:     interval,
	}
}

// Start spawns a non-blocking background ticker loop to monitor accounts
func (sc *SuspensionChecker) Start(ctx context.Context) {
	ticker := time.NewTicker(sc.interval)
	go func() {
		fmt.Printf("[Worker] Automated Suspension Checker started. Running scans every %v\n", sc.interval)
		for {
			select {
			case <-ticker.C:
				if err := sc.ScanAndSuspendOverdue(ctx); err != nil {
					fmt.Printf("[Worker Error] Failed execution during suspension check loop: %v\n", err)
				}
			case <-ctx.Done():
				ticker.Stop()
				fmt.Println("[Worker] Suspension Checker stopped gracefully.")
				return
			}
		}
	}()
}

// ScanAndSuspendOverdue cross-references past-due invoices with live PPPoE connections
func (sc *SuspensionChecker) ScanAndSuspendOverdue(ctx context.Context) error {
	// 1. Query all unpaid invoices that are past their grace period due-dates
	overdueInvoices, err := sc.invoiceRepo.ListOverdue(ctx, time.Now())
	if err != nil {
		return fmt.Errorf("failed to pull overdue accounts: %w", err)
	}

	if len(overdueInvoices) == 0 {
		return nil
	}

	fmt.Printf("[Worker] Found %d overdue invoices. Processing isolation rules...\n", len(overdueInvoices))

	for _, invoice := range overdueInvoices {
		// 2. Resolve client parameters
		customer, err := sc.customerRepo.GetByID(ctx, invoice.CustomerID)
		if err != nil {
			fmt.Printf("[Worker Error] Skipping invoice %s; customer lookup failed: %v\n", invoice.Reference, err)
			continue
		}

		// Skip if the customer is already marked as suspended in the database
		if customer.Status == "suspended" {
			continue
		}

		// 3. Trigger active router disconnection commands
		fmt.Printf("[Worker] Suspending customer account %s on router due to unpaid invoice %s\n", customer.MikrotikUser, invoice.Reference)
		if err := sc.provisioner.SuspendUser(ctx, customer); err != nil {
			fmt.Printf("[Worker Error] Failed router disconnection command for user %s: %v\n", customer.MikrotikUser, err)
			continue
		}

		// 4. Record state mutation to database ledger profiles
		customer.Status = "suspended"
		customer.UpdatedAt = time.Now()
		if err := sc.customerRepo.Update(ctx, customer); err != nil {
			fmt.Printf("[Worker Error] Failed to update customer status flag in DB: %v\n", err)
		}
	}

	return nil
}
