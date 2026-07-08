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

func (r *InvoiceRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, paidAt *string) error {
	query := `
		UPDATE invoices 
		SET status = $1, paid_at = $2, updated_at = $3 
		WHERE id = $4
	`
	// Fallback conversion parsing to allow strings to cleanly flow into your timestamp columns
	var parsedTime interface{} = nil
	if paidAt != nil && *paidAt != "" {
		t, err := time.Parse(time.RFC3339, *paidAt)
		if err == nil {
			parsedTime = t
		} else {
			parsedTime = *paidAt // If it's already formatting-compliant string layout
		}
	}

	_, err := r.db.ExecContext(ctx, query, status, parsedTime, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update state parameters for invoice record: %w", err)
	}
	return nil
}

// Add these methods at the bottom of your invoice_repo.go file

// MarkAsPaid sets an invoice status explicitly to "paid".
func (r *InvoiceRepository) MarkAsPaid(ctx context.Context, id uuid.UUID, paidAt time.Time) error {
	query := `
		UPDATE invoices 
		SET status = 'paid', paid_at = $1, updated_at = $2 
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, paidAt, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to mark invoice as paid: %w", err)
	}
	return nil
}

// ListOverdue fetches all invoices that remain unpaid past their due date constraint.
func (r *InvoiceRepository) ListOverdue(ctx context.Context, asOf time.Time) ([]*domain.Invoice, error) {
	query := `
		SELECT id, customer_id, number, amount, tax, total, status, description, period_start, period_end, due_date, paid_at, pdf_url, created_at, updated_at
		FROM invoices 
		WHERE status IN ('pending', 'unpaid') AND due_date < $1
		ORDER BY due_date ASC
	`
	rows, err := r.db.QueryContext(ctx, query, asOf)
	if err != nil {
		return nil, fmt.Errorf("failed to query overdue invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*domain.Invoice
	for rows.Next() {
		var inv domain.Invoice
		err := rows.Scan(
			&inv.ID, &inv.CustomerID, &inv.Number, &inv.Amount, &inv.Tax, &inv.Total,
			&inv.Status, &inv.Description, &inv.PeriodStart, &inv.PeriodEnd, &inv.DueDate,
			&inv.PaidAt, &inv.PDFURL, &inv.CreatedAt, &inv.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan overdue invoice row: %w", err)
		}
		invoices = append(invoices, &inv)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return invoices, nil
}

func (r *InvoiceRepository) GetByReference(ctx context.Context, reference string) (*domain.Invoice, error) {
	query := `
		SELECT id, customer_id, number, amount, tax, total, status, description, period_start, period_end, due_date, paid_at, pdf_url, created_at, updated_at
		FROM invoices WHERE number = $1
	`
	var inv domain.Invoice
	err := r.db.QueryRowContext(ctx, query, reference).Scan(
		&inv.ID, &inv.CustomerID, &inv.Number, &inv.Amount, &inv.Tax, &inv.Total,
		&inv.Status, &inv.Description, &inv.PeriodStart, &inv.PeriodEnd, &inv.DueDate,
		&inv.PaidAt, &inv.PDFURL, &inv.CreatedAt, &inv.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch invoice by reference number: %w", err)
	}
	return &inv, nil
}

// Add this method to internal/repository/postgres/invoice_repo.go

func (r *InvoiceRepository) Update(ctx context.Context, inv *domain.Invoice) error {
	query := `
		UPDATE invoices 
		SET customer_id = $1, number = $2, amount = $3, tax = $4, total = $5, 
		    status = $6, description = $7, period_start = $8, period_end = $9, 
		    due_date = $10, paid_at = $11, pdf_url = $12, updated_at = $13 
		WHERE id = $14
	`
	_, err := r.db.ExecContext(ctx, query,
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
		inv.PaidAt,
		inv.PDFURL,
		time.Now(),
		inv.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update invoice record: %w", err)
	}
	return nil
}

// ExistsByReference checks for duplicate reference identifiers across the persistence layer.
func (r *InvoiceRepository) ExistsByReference(ctx context.Context, reference string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM invoices WHERE number = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, reference).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check invoice structural duplication: %w", err)
	}
	return exists, nil
}
