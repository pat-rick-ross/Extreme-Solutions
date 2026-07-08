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

// Add this method to internal/repository/postgres/payment_repo.go

func (r *PaymentRepository) ListByCustomerID(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]*domain.Payment, int64, error) {
	query := `
		SELECT id, invoice_id, customer_id, amount, status, reference, provider, created_at, updated_at, COUNT(*) OVER()
		FROM payments
		WHERE customer_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, customerID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query customer payment logs: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment
	var totalCount int64 = 0

	for rows.Next() {
		var p domain.Payment
		err := rows.Scan(
			&p.ID,
			&p.InvoiceID,
			&p.CustomerID,
			&p.Amount,
			&p.Status,
			&p.Reference,
			&p.Provider,
			&p.CreatedAt,
			&p.UpdatedAt,
			&totalCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan payment record row: %w", err)
		}
		payments = append(payments, &p)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	return payments, totalCount, nil
}

// Add this method to internal/repository/postgres/payment_repo.go

func (r *PaymentRepository) ListPending(ctx context.Context) ([]*domain.Payment, error) {
	query := `
		SELECT id, invoice_id, customer_id, amount, status, reference, provider, created_at, updated_at
		FROM payments
		WHERE status = 'pending'
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending transaction ledger logs: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		var p domain.Payment
		err := rows.Scan(
			&p.ID,
			&p.InvoiceID,
			&p.CustomerID,
			&p.Amount,
			&p.Status,
			&p.Reference,
			&p.Provider,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pending payment dataset row: %w", err)
		}
		payments = append(payments, &p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return payments, nil
}

// Add this method to internal/repository/postgres/payment_repo.go

func (r *PaymentRepository) Update(ctx context.Context, p *domain.Payment) error {
	query := `
		UPDATE payments 
		SET invoice_id = $1, customer_id = $2, amount = $3, status = $4, 
		    reference = $5, provider = $6, updated_at = $7
		WHERE id = $8
	`
	_, err := r.db.ExecContext(ctx, query,
		p.InvoiceID,
		p.CustomerID,
		p.Amount,
		p.Status,
		p.Reference,
		p.Provider,
		time.Now(),
		p.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update payment record logs: %w", err)
	}
	return nil
}

// Add this method to internal/repository/postgres/payment_repo.go

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE payments 
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update payment record status: %w", err)
	}
	return nil
}

// Add this method to internal/repository/postgres/payment_repo.go

func (r *PaymentRepository) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	query := `
		SELECT id, invoice_id, customer_id, amount, status, reference, provider, created_at, updated_at
		FROM payments 
		WHERE id = $1
	`
	// Handle parsing safely inside the database execution node boundaries
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid string query identity representation passed to transaction engine: %w", err)
	}

	var p domain.Payment
	err = r.db.QueryRowContext(ctx, query, parsedID).Scan(
		&p.ID,
		&p.InvoiceID,
		&p.CustomerID,
		&p.Amount,
		&p.Status,
		&p.Reference,
		&p.Provider,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payment by structural transaction id lookup: %w", err)
	}
	return &p, nil
}

func (r *PaymentRepository) CompletePayment(ctx context.Context, id uuid.UUID, reference string) error {
	query := `
		UPDATE payments 
		SET status = 'completed', reference = $1, updated_at = $2 
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, reference, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to execute payment completion transaction: %w", err)
	}
	return nil
}
