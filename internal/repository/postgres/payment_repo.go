package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/isp-billing/internal/domain"
)

type paymentRepository struct {
	db *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) *paymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	if payment.ID == uuid.Nil {
		payment.ID = uuid.New()
	}

	query := `
		INSERT INTO payments (id, invoice_id, customer_id, amount, status, method, provider, reference, mpesa_receipt, mpesa_phone, description, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	now := time.Now()
	payment.CreatedAt = now
	payment.UpdatedAt = now

	if payment.Status == "" {
		payment.Status = domain.PaymentStatusPending
	}

	_, err := r.db.Exec(ctx, query,
		payment.ID,
		payment.InvoiceID,
		payment.CustomerID,
		payment.Amount,
		payment.Status,
		payment.Method,
		payment.Provider,
		payment.Reference,
		payment.MpesaReceipt,
		payment.MpesaPhone,
		payment.Description,
		payment.Metadata,
		payment.CreatedAt,
		payment.UpdatedAt,
	)

	return err
}

func (r *paymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	query := `
		SELECT id, invoice_id, customer_id, amount, status, method, provider, reference, mpesa_receipt, mpesa_phone, description, metadata, completed_at, created_at, updated_at
		FROM payments
		WHERE id = $1
	`

	var payment domain.Payment
	err := r.db.QueryRow(ctx, query, id).Scan(
		&payment.ID,
		&payment.InvoiceID,
		&payment.CustomerID,
		&payment.Amount,
		&payment.Status,
		&payment.Method,
		&payment.Provider,
		&payment.Reference,
		&payment.MpesaReceipt,
		&payment.MpesaPhone,
		&payment.Description,
		&payment.Metadata,
		&payment.CompletedAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return &payment, nil
}

func (r *paymentRepository) GetByReference(ctx context.Context, reference string) (*domain.Payment, error) {
	query := `
		SELECT id, invoice_id, customer_id, amount, status, method, provider, reference, mpesa_receipt, mpesa_phone, description, metadata, completed_at, created_at, updated_at
		FROM payments
		WHERE reference = $1
	`

	var payment domain.Payment
	err := r.db.QueryRow(ctx, query, reference).Scan(
		&payment.ID,
		&payment.InvoiceID,
		&payment.CustomerID,
		&payment.Amount,
		&payment.Status,
		&payment.Method,
		&payment.Provider,
		&payment.Reference,
		&payment.MpesaReceipt,
		&payment.MpesaPhone,
		&payment.Description,
		&payment.Metadata,
		&payment.CompletedAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get payment by reference: %w", err)
	}

	return &payment, nil
}

func (r *paymentRepository) GetByInvoiceID(ctx context.Context, invoiceID uuid.UUID) (*domain.Payment, error) {
	query := `
		SELECT id, invoice_id, customer_id, amount, status, method, provider, reference, mpesa_receipt, mpesa_phone, description, metadata, completed_at, created_at, updated_at
		FROM payments
		WHERE invoice_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var payment domain.Payment
	err := r.db.QueryRow(ctx, query, invoiceID).Scan(
		&payment.ID,
		&payment.InvoiceID,
		&payment.CustomerID,
		&payment.Amount,
		&payment.Status,
		&payment.Method,
		&payment.Provider,
		&payment.Reference,
		&payment.MpesaReceipt,
		&payment.MpesaPhone,
		&payment.Description,
		&payment.Metadata,
		&payment.CompletedAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get payment by invoice: %w", err)
	}

	return &payment, nil
}

func (r *paymentRepository) GetByMpesaReceipt(ctx context.Context, receipt string) (*domain.Payment, error) {
	query := `
		SELECT id, invoice_id, customer_id, amount, status, method, provider, reference, mpesa_receipt, mpesa_phone, description, metadata, completed_at, created_at, updated_at
		FROM payments
		WHERE mpesa_receipt = $1
	`

	var payment domain.Payment
	err := r.db.QueryRow(ctx, query, receipt).Scan(
		&payment.ID,
		&payment.InvoiceID,
		&payment.CustomerID,
		&payment.Amount,
		&payment.Status,
		&payment.Method,
		&payment.Provider,
		&payment.Reference,
		&payment.MpesaReceipt,
		&payment.MpesaPhone,
		&payment.Description,
		&payment.Metadata,
		&payment.CompletedAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get payment by receipt: %w", err)
	}

	return &payment, nil
}

func (r *paymentRepository) ListByCustomerID(ctx context.Context, customerID uuid.UUID, page, pageSize int) ([]*domain.Payment, int64, error) {
	countQuery := `SELECT COUNT(*) FROM payments WHERE customer_id = $1`
	var total int64
	err := r.db.QueryRow(ctx, countQuery, customerID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count payments: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id, invoice_id, customer_id, amount, status, method, provider, reference, mpesa_receipt, mpesa_phone, description, metadata, completed_at, created_at, updated_at
		FROM payments
		WHERE customer_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, customerID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list payments: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		var payment domain.Payment
		err := rows.Scan(
			&payment.ID,
			&payment.InvoiceID,
			&payment.CustomerID,
			&payment.Amount,
			&payment.Status,
			&payment.Method,
			&payment.Provider,
			&payment.Reference,
			&payment.MpesaReceipt,
			&payment.MpesaPhone,
			&payment.Description,
			&payment.Metadata,
			&payment.CompletedAt,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan payment: %w", err)
		}
		payments = append(payments, &payment)
	}

	return payments, total, nil
}

func (r *paymentRepository) ListPending(ctx context.Context) ([]*domain.Payment, error) {
	query := `
		SELECT id, invoice_id, customer_id, amount, status, method, provider, reference, mpesa_receipt, mpesa_phone, description, metadata, completed_at, created_at, updated_at
		FROM payments
		WHERE status = $1 AND created_at < $2
	`

	timeout := time.Now().Add(-5 * time.Minute)
	rows, err := r.db.Query(ctx, query, domain.PaymentStatusPending, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending payments: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		var payment domain.Payment
		err := rows.Scan(
			&payment.ID,
			&payment.InvoiceID,
			&payment.CustomerID,
			&payment.Amount,
			&payment.Status,
			&payment.Method,
			&payment.Provider,
			&payment.Reference,
			&payment.MpesaReceipt,
			&payment.MpesaPhone,
			&payment.Description,
			&payment.Metadata,
			&payment.CompletedAt,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}
		payments = append(payments, &payment)
	}

	return payments, nil
}

func (r *paymentRepository) Update(ctx context.Context, payment *domain.Payment) error {
	query := `
		UPDATE payments
		SET amount = $1, status = $2, method = $3, provider = $4, reference = $5, mpesa_receipt = $6, mpesa_phone = $7, description = $8, metadata = $9, completed_at = $10, updated_at = $11
		WHERE id = $12
	`

	payment.UpdatedAt = time.Now()

	_, err := r.db.Exec(ctx, query,
		payment.Amount,
		payment.Status,
		payment.Method,
		payment.Provider,
		payment.Reference,
		payment.MpesaReceipt,
		payment.MpesaPhone,
		payment.Description,
		payment.Metadata,
		payment.CompletedAt,
		payment.UpdatedAt,
		payment.ID,
	)

	return err
}

func (r *paymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus) error {
	query := `UPDATE payments SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(ctx, query, status, time.Now(), id)
	return err
}

func (r *paymentRepository) CompletePayment(ctx context.Context, id uuid.UUID, receipt string) error {
	now := time.Now()
	query := `
		UPDATE payments 
		SET status = $1, mpesa_receipt = $2, completed_at = $3, updated_at = $4 
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query, domain.PaymentStatusCompleted, receipt, now, now, id)
	return err
}

func (r *paymentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM payments WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
