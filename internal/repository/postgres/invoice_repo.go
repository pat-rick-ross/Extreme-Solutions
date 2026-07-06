package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"Extreme-Solutions/internal/domain"

	"github.com/google/uuid"
)

type InvoiceRepository struct {
	db *sql.DB
}

func NewInvoiceRepository(db *sql.DB) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

func (r *InvoiceRepository) Create(ctx context.Context, inv *domain.Invoice) error {
	query := `
		INSERT INTO invoices (id, customer_id, number, amount, tax, total, status, description, period_start, period_end, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	if inv.ID == uuid.Nil {
		inv.ID = uuid.New()
	}
	now := time.Now()
	inv.CreatedAt = now
	inv.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		inv.ID,
		inv.CustomerID,
		inv.Number,
		inv.Amount,
		inv.Tax,
		inv.Total,
		inv.Status,
		inv.Description,
		inv.PeriodStart,
		inv.PeriodEnd,
		inv.DueDate,
		inv.CreatedAt,
		inv.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}
	return nil
}

func (r *InvoiceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	query := `
		SELECT id, customer_id, number, amount, tax, total, status, description, period_start, period_end, due_date, paid_at, pdf_url, created_at, updated_at
		FROM invoices WHERE id = $1
	`
	var inv domain.Invoice
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&inv.ID, &inv.CustomerID, &inv.Number, &inv.Amount, &inv.Tax, &inv.Total,
		&inv.Status, &inv.Description, &inv.PeriodStart, &inv.PeriodEnd, &inv.DueDate,
		&inv.PaidAt, &inv.PDFURL, &inv.CreatedAt, &inv.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch invoice by id: %w", err)
	}
	return &inv, nil
}

func (r *InvoiceRepository) GetUnpaidByCustomerID(ctx context.Context, customerID uuid.UUID) ([]domain.Invoice, error) {
	query := `
		SELECT id, customer_id, number, amount, tax, total, status, description, period_start, period_end, due_date, paid_at, pdf_url, created_at, updated_at
		FROM invoices 
		WHERE customer_id = $1 AND status IN ('pending', 'unpaid')
		ORDER BY due_date ASC
	`
	rows, err := r.db.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query unpaid invoices: %w", err)
	}
	defer rows.Close()

	var invoices []domain.Invoice
	for rows.Next() {
		var inv domain.Invoice
		err := rows.Scan(
			&inv.ID, &inv.CustomerID, &inv.Number, &inv.Amount, &inv.Tax, &inv.Total,
			&inv.Status, &inv.Description, &inv.PeriodStart, &inv.PeriodEnd, &inv.DueDate,
			&inv.PaidAt, &inv.PDFURL, &inv.CreatedAt, &inv.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice row: %w", err)
		}
		invoices = append(invoices, inv)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return invoices, nil
}

func (r *InvoiceRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, paidAt *time.Time) error {
	query := `
		UPDATE invoices 
		SET status = $1, paid_at = $2, updated_at = $3 
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query, status, paidAt, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}
	return nil
}