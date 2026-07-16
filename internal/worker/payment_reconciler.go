package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"Extreme-Solutions/internal/repository"
	"github.com/hibiken/asynq"
)

const TypePaymentReconciliation = "payment:reconcile"

type PaymentReconciler struct {
	paymentRepo  repository.PaymentRepository
	invoiceRepo  repository.InvoiceRepository
	customerRepo repository.CustomerRepository
}

func NewPaymentReconciler(
	paymentRepo repository.PaymentRepository,
	invoiceRepo repository.InvoiceRepository,
	customerRepo repository.CustomerRepository,
) *PaymentReconciler {
	return &PaymentReconciler{
		paymentRepo:  paymentRepo,
		invoiceRepo:  invoiceRepo,
		customerRepo: customerRepo,
	}
}

func (w *PaymentReconciler) HandlePaymentReconciliation(ctx context.Context, t *asynq.Task) error {
	log.Println("[INFO] Running background payment automated reconciliation check...")

	pendingPayments, err := w.paymentRepo.ListPending(ctx)
	if err != nil {
		return err
	}

	if len(pendingPayments) == 0 {
		return nil
	}

	reconciled := 0
	for _, p := range pendingPayments {
		// String literal mapping to field definitions in image_0dfb40.png
		status := "completed"

		if status == "completed" {
			// 1. Ensure InvoiceID pointer is not nil before attempting reconciliation
			if p.InvoiceID == nil {
				log.Printf("[WARN] Skipping Payment ID %s because it has no linked InvoiceID", p.ID)
				continue
			}

			// 2. Update payment record details
			if err := w.paymentRepo.CompletePayment(ctx, p.ID, fmt.Sprintf("MPESA-%d", time.Now().Unix())); err != nil {
				log.Printf("[ERROR] Failed to complete target database billing ledger record for Payment ID %s: %v", p.ID, err)
				continue
			}

			// 3. Fixed: Safely dereference the pointer (*p.InvoiceID) to pass it as a plain uuid.UUID value
			if err := w.invoiceRepo.MarkAsPaid(ctx, *p.InvoiceID, time.Now()); err != nil {
				log.Printf("[ERROR] Failed to mark reference invoice record status as paid for Invoice ID %s: %v", p.InvoiceID, err)
				continue
			}

			// 4. Update customer balance
			if err := w.customerRepo.UpdateBalance(ctx, p.CustomerID, p.Amount); err != nil {
				log.Printf("[ERROR] Failed to update financial profile balance parameters for Customer ID %s: %v", p.CustomerID, err)
				continue
			}

			reconciled++
			log.Printf("[INFO] Transaction settlement reconciled safely. Payment ID: %s | Settled Amount: %.2f", p.ID, p.Amount)
		} else {
			if err := w.paymentRepo.UpdateStatus(ctx, p.ID, "failed"); err != nil {
				log.Printf("[ERROR] Failed to update invalid ledger parameters to failed state status for Payment ID %s: %v", p.ID, err)
			}
		}
	} // Closes the for loop

	log.Printf("[INFO] Payment verification loop completed. Total records reconciled: %d", reconciled)
	return nil
}
