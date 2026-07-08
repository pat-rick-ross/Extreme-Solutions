package payment

import (
	"context"
	"fmt"
	"time"

	"Extreme-Solutions/internal/domain"
	"Extreme-Solutions/internal/repository"
	"Extreme-Solutions/internal/service/network"
)

type PaymentReconciler struct {
	paymentRepo  repository.PaymentRepository
	invoiceRepo  repository.InvoiceRepository
	customerRepo repository.CustomerRepository
	provisioner  *network.Provisioner
}

func NewPaymentReconciler(
	pr repository.PaymentRepository,
	ir repository.InvoiceRepository,
	cr repository.CustomerRepository,
	provisioner *network.Provisioner,
) *PaymentReconciler {
	return &PaymentReconciler{
		paymentRepo:  pr,
		invoiceRepo:  ir,
		customerRepo: cr,
		provisioner:  provisioner,
	}
}

// ProcessSuccessfulPayment safely updates accounting states and provisions hardware access rules
func (r *PaymentReconciler) ProcessSuccessfulPayment(ctx context.Context, payment *domain.Payment, invoiceRef string) error {
	// 1. Persist payment record to database layer to maintain clean ledger logs
	if err := r.paymentRepo.Create(ctx, payment); err != nil {
		return fmt.Errorf("failed to log database payment ledger record: %w", err)
	}

	// 2. Resolve target Invoice details via its unique tracking reference
	invoice, err := r.invoiceRepo.GetByReference(ctx, invoiceRef)
	if err != nil {
		return fmt.Errorf("failed to locate invoice reference %s: %w", invoiceRef, err)
	}

	// 3. Perform mutation logic on financial states
	invoice.Status = "paid"
	invoice.UpdatedAt = time.Now()
	if err := r.invoiceRepo.Update(ctx, invoice); err != nil {
		return fmt.Errorf("failed to transition invoice status to paid: %w", err)
	}

	// 4. Load Customer account data profiles to extract MikroTik user tags
	customer, err := r.customerRepo.GetByID(ctx, invoice.CustomerID)
	if err != nil {
		return fmt.Errorf("payment complete but target customer record resolution failed: %w", err)
	}

	// 5. AUTOMATION HANDSHAKE: Provision network line profile instantly on router hardware
	fmt.Printf("[Reconciler] Payment received for customer %s. Initiating router provisioning thread...\n", customer.ID)
	if err := r.provisioner.ReactivateUser(ctx, customer); err != nil {
		// Log error but do not rollback financial records—the user paid, the network state just needs manual retry sync
		fmt.Printf("[CRITICAL NETWORK ERROR] Failed to reactivate customer line interface on router: %v\n", err)
		return fmt.Errorf("payment settled successfully but router provisioning timed out: %w", err)
	}

	fmt.Printf("[✓ SUCCESS] Billing status updated and network interface active for subscriber: %s\n", customer.MikrotikUser)
	return nil
}
