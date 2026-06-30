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
	logger.Info("Running payment reconciliation", nil)

	// Get pending payments older than 5 minutes
	pendingPayments, err := w.paymentRepo.ListPending(ctx)
	if err != nil {
		return err
	}

	if len(pendingPayments) == 0 {
		return nil
	}

	reconciled := 0
	for _, payment := range pendingPayments {
		// Here you would call M-PESA API to check status
		// For now, we'll simulate a completed payment
		status := domain.PaymentStatusCompleted

		if status == domain.PaymentStatusCompleted {
			// Update payment
			if err := w.paymentRepo.CompletePayment(ctx, payment.ID, fmt.Sprintf("MPESA-%d", time.Now().Unix())); err != nil {
				logger.Error("Failed to complete payment", map[string]interface{}{
					"payment_id": payment.ID,
					"error":      err,
				})
				continue
			}

			// Mark invoice as paid
			if err := w.invoiceRepo.MarkAsPaid(ctx, payment.InvoiceID, time.Now()); err != nil {
				logger.Error("Failed to mark invoice as paid", map[string]interface{}{
					"invoice_id": payment.InvoiceID,
					"error":      err,
				})
				continue
			}

			// Update customer balance
			if err := w.customerRepo.UpdateBalance(ctx, payment.CustomerID, payment.Amount); err != nil {
				logger.Error("Failed to update customer balance", map[string]interface{}{
					"customer_id": payment.CustomerID,
					"error":       err,
				})
				continue
			}

			reconciled++
			logger.Info("Payment reconciled", map[string]interface{}{
				"payment_id": payment.ID,
				"amount":     payment.Amount,
			})
		} else {
			// Mark payment as failed
			if err := w.paymentRepo.UpdateStatus(ctx, payment.ID, domain.PaymentStatusFailed); err != nil {
				logger.Error("Failed to mark payment as failed", map[string]interface{}{
					"payment_id": payment.ID,
					"error":      err,
				})
			}
		}
	}

	logger.Info("Payment reconciliation completed", map[string]interface{}{
		"reconciled": reconciled,
	})

	return nil
}
