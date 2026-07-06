package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"Extreme-Solutions/internal/domain"

	"github.com/google/uuid"
)

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	query := `
		INSERT INTO payments (id, invoice_id, customer_id, amount, status, method, provider, reference, mpesa_receipt, mpesa_phone, description, metadata, completed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		p.ID,
		p.InvoiceID,
		p.CustomerID,
		p.Amount,
		p.Status,
		p.Method,
		p.Provider,
		p.Reference,
		p.MpesaReceipt,
		p.MpesaPhone,
		p.Description,
		p.Metadata, // Passes cleanly as standard raw json bytes into Postgres JSONB
		p.CompletedAt,
		p.CreatedAt,
		p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert payment execution data: %w", err)
	}
	return nil
}

func (r *PaymentRepository) GetByReference(ctx context.Context, reference string) (*domain.Payment, error) {
	query := `
		SELECT id, invoice_id, customer_id, amount, status, method, provider, reference, mpesa_receipt, mpesa_phone, description, metadata, completed_at, created_at, updated_at
		FROM payments 
		WHERE reference = $1
	`
	var p domain.Payment
	err := r.db.QueryRowContext(ctx, query, reference).Scan(
		&p.ID, &p.InvoiceID, &p.CustomerID, &p.Amount, &p.Status, &p.Method, &p.Provider,
		&p.Reference, &p.MpesaReceipt, &p.MpesaPhone, &p.Description, &p.Metadata,
		&p.CompletedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to track down payment by reference code: %w", err)
	}
	return &p, nil
}