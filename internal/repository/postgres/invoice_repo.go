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

type invoiceRepository struct {
	db *pgxpool.Pool
}

func NewInvoiceRepository(db *pgxpool.Pool) *invoiceRepository {
	return &invoiceRepository{db: db}
}

func (r *invoiceRepository) Create(ctx context.Context, invoice *domain.Invoice) error {
	if invoice.ID == uuid.Nil {
		invoice.ID = uuid.New()
	}

	query := `
		INSERT INTO invoices (id, customer_id, number, amount, tax, total, status, description, period_start, period_end, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	now := time.Now()
	invoice.CreatedAt = now
	invoice.UpdatedAt = now

	if invoice.Status == "" {
		invoice.Status = domain.InvoiceStatusPending
	}

	_, err := r.db.Exec(ctx, query,
		invoice.ID,
		invoice.CustomerID,
		invoice.Number,
		invoice.Amount,
		invoice.Tax,
		invoice.Total,
		invoice.Status,
		invoice.Description,
		invoice.PeriodStart,
		invoice.PeriodEnd,
		invoice.DueDate,
		invoice.CreatedAt,
		invoice.UpdatedAt,
	)

	return err
}

func (r *invoiceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	query := `
		SELECT id, customer_id, number, amount, tax, total, status, description, period_start, period_end, due_date, paid_at, pdf_url, created_at, updated_at
		FROM invoices
		WHERE id = $1
	`

	var invoice domain.Invoice
	err := r.db.QueryRow(ctx, query, id).Scan(
		&invoice.ID,
		&invoice.CustomerID,
		&invoice.Number,
		&invoice.Amount,
		&invoice.Tax,
		&invoice.Total,
		&invoice.Status,
		&invoice.Description,
		&invoice.PeriodStart,
		&invoice.PeriodEnd,
		&invoice.DueDate,
		&invoice.PaidAt,
		&invoice.PDFURL,
		&invoice.CreatedAt,
		&invoice.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return &invoice, nil
}

func (r *invoiceRepository) GetByNumber(ctx context.Context, number string) (*domain.Invoice, error) {
	query := `
		SELECT id, customer_id, number, amount, tax, total, status, description, period_start, period_end, due_date, paid_at, pdf_url, created_at, updated_at
		FROM invoices
		WHERE number = $1
	`

	var invoice domain.Invoice
	err := r.db.QueryRow(ctx, query, number).Scan(
		&invoice.ID,
		&invoice.CustomerID,
		&invoice.Number,
		&invoice.Amount,
		&invoice.Tax,
		&invoice.Total,
		&invoice.Status,
		&invoice.Description,
		&invoice.PeriodStart,
		&invoice.PeriodEnd,
		&invoice.DueDate,
		&invoice.PaidAt,
		&invoice.PDFURL,
		&invoice.CreatedAt,
		&invoice.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get invoice by number: %w", err)
	}

	return &invoice, nil
}

func (r *invoiceRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID, page, pageSize int) ([]*domain.Invoice, int64, error) {
	countQuery := `SELECT COUNT(*) FROM invoices WHERE customer_id = $1`
	var total int64
	err := r.db.QueryRow(ctx, countQuery, customerID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id, customer_id, number, amount, tax, total, status, description, period_start, period_end, due_date, paid_at, pdf_url, created_at, updated_at
		FROM invoices
		WHERE customer_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, customerID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get invoices by customer: %w", err)
	}
	defer rows.Close()

	var invoices []*domain.Invoice
	for rows.Next() {
		var invoice domain.Invoice
		err := rows.Scan(
			&invoice.ID,
			&invoice.CustomerID,
			&invoice.Number,
			&invoice.Amount,
			&invoice.Tax,
			&invoice.Total,
			&invoice.Status,
			&invoice.Description,
			&invoice.PeriodStart,
			&invoice.PeriodEnd,
			&invoice.DueDate,
			&invoice.PaidAt,
			&invoice.PDFURL,
			&invoice.CreatedAt,
			&invoice.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, &invoice)
	}

	return invoices, total, nil
}

func (r *invoiceRepository) List(ctx context.Context, filter *domain.InvoiceFilter) ([]*domain.Invoice, int64, error) {
	// Implementation similar to customer list with filters
	// For brevity, simplified version
	query := `
		SELECT id, customer_id, number, amount, tax, total, status, description, period_start, period_end, due_date, paid_at, pdf_url, created_at, updated_at
		FROM invoices
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	offset := (filter.Page - 1) * filter.PageSize
	rows, err := r.db.Query(ctx, query, filter.PageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*domain.Invoice
	for rows.Next() {
		var invoice domain.Invoice
		err := rows.Scan(
			&invoice.ID,
			&invoice.CustomerID,
			&invoice.Number,
			&invoice.Amount,
			&invoice.Tax,
			&invoice.Total,
			&invoice.Status,
			&invoice.Description,
			&invoice.PeriodStart,
			&invoice.PeriodEnd,
			&invoice.DueDate,
			&invoice.PaidAt,
			&invoice.PDFURL,
			&invoice.CreatedAt,
			&invoice.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, &invoice)
	}

	return invoices, int64(len(invoices)), nil
}

func (r *invoiceRepository) ListPending(ctx context.Context) ([]*domain.Invoice, error) {
	query := `
		SELECT id, customer_id, number, amount, tax, total, status, description, period_start, period_end, due_date, paid_at, pdf_url, created_at, updated_at
		FROM invoices
		WHERE status = $1
	`

	rows, err := r.db.Query(ctx, query, domain.InvoiceStatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*domain.Invoice
	for rows.Next() {
		var invoice domain.Invoice
		err := rows.Scan(
			&invoice.ID,
			&invoice.CustomerID,
			&invoice.Number,
			&invoice.Amount,
			&invoice.Tax,
			&invoice.Total,
			&invoice.Status,
			&invoice.Description,
			&invoice.PeriodStart,
			&invoice.PeriodEnd,
			&invoice.DueDate,
			&invoice.PaidAt,
			&invoice.PDFURL,
			&invoice.CreatedAt,
			&invoice.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, &invoice)
	}

	return invoices, nil
}

func (r *invoiceRepository) ListOverdue(ctx context.Context) ([]*domain.Invoice, error) {
	query := `
		SELECT id, customer_id, number, amount, tax, total, status, description, period_start, period_end, due_date, paid_at, pdf_url, created_at, updated_at
		FROM invoices
		WHERE status = $1 AND due_date < $2
	`

	rows, err := r.db.Query(ctx, query, domain.InvoiceStatusPending, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to list overdue invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*domain.Invoice
	for rows.Next() {
		var invoice domain.Invoice
		err := rows.Scan(
			&invoice.ID,
			&invoice.CustomerID,
			&invoice.Number,
			&invoice.Amount,
			&invoice.Tax,
			&invoice.Total,
			&invoice.Status,
			&invoice.Description,
			&invoice.PeriodStart,
			&invoice.PeriodEnd,
			&invoice.DueDate,
			&invoice.PaidAt,
			&invoice.PDFURL,
			&invoice.CreatedAt,
			&invoice.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, &invoice)
	}

	return invoices, nil
}

func (r *invoiceRepository) Update(ctx context.Context, invoice *domain.Invoice) error {
	query := `
		UPDATE invoices
		SET amount = $1, tax = $2, total = $3, status = $4, description = $5, period_start = $6, period_end = $7, due_date = $8, pdf_url = $9, updated_at = $10
		WHERE id = $11
	`

	invoice.UpdatedAt = time.Now()

	_, err := r.db.Exec(ctx, query,
		invoice.Amount,
		invoice.Tax,
		invoice.Total,
		invoice.Status,
		invoice.Description,
		invoice.PeriodStart,
		invoice.PeriodEnd,
		invoice.DueDate,
		invoice.PDFURL,
		invoice.UpdatedAt,
		invoice.ID,
	)

	return err
}

func (r *invoiceRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.InvoiceStatus) error {
	query := `UPDATE invoices SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(ctx, query, status, time.Now(), id)
	return err
}

func (r *invoiceRepository) MarkAsPaid(ctx context.Context, id uuid.UUID, paidAt time.Time) error {
	query := `UPDATE invoices SET status = $1, paid_at = $2, updated_at = $3 WHERE id = $4`
	_, err := r.db.Exec(ctx, query, domain.InvoiceStatusPaid, paidAt, time.Now(), id)
	return err
}

func (r *invoiceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM invoices WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
