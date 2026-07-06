package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"Extreme-Solutions/internal/domain"

	"github.com/google/uuid"
)

type CustomerRepository struct {
	db *sql.DB
}

// NewCustomerRepository initializes the postgres implementation
func NewCustomerRepository(db *sql.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) Create(ctx context.Context, customer *domain.Customer) error {
	query := `
		INSERT INTO customers (id, first_name, last_name, email, phone, address, package_id, balance, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	if customer.ID == uuid.Nil {
		customer.ID = uuid.New()
	}
	now := time.Now()
	customer.CreatedAt = now
	customer.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		customer.ID,
		customer.FirstName,
		customer.LastName,
		customer.Email,
		customer.Phone,
		customer.Address,
		customer.PackageID,
		customer.Balance,
		customer.Status,
		customer.CreatedAt,
		customer.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert customer: %w", err)
	}
	return nil
}

func (r *CustomerRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, address, package_id, balance, status, created_at, updated_at, suspended_at, last_login_at
		FROM customers WHERE id = $1
	`
	var c domain.Customer
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.FirstName, &c.LastName, &c.Email, &c.Phone, &c.Address,
		&c.PackageID, &c.Balance, &c.Status, &c.CreatedAt, &c.UpdatedAt,
		&c.SuspendedAt, &c.LastLoginAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // Return nil, nil when no records match cleanly
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch customer by id: %w", err)
	}
	return &c, nil
}

func (r *CustomerRepository) GetByPhone(ctx context.Context, phone string) (*domain.Customer, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, address, package_id, balance, status, created_at, updated_at, suspended_at, last_login_at
		FROM customers WHERE phone = $1
	`
	var c domain.Customer
	err := r.db.QueryRowContext(ctx, query, phone).Scan(
		&c.ID, &c.FirstName, &c.LastName, &c.Email, &c.Phone, &c.Address,
		&c.PackageID, &c.Balance, &c.Status, &c.CreatedAt, &c.UpdatedAt,
		&c.SuspendedAt, &c.LastLoginAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch customer by phone: %w", err)
	}
	return &c, nil
}

func (r *CustomerRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	var err error
	if status == "suspended" {
		query := `UPDATE customers SET status = $1, suspended_at = $2, updated_at = $3 WHERE id = $4`
		_, err = r.db.ExecContext(ctx, query, status, time.Now(), time.Now(), id)
	} else {
		query := `UPDATE customers SET status = $1, suspended_at = NULL, updated_at = $2 WHERE id = $3`
		_, err = r.db.ExecContext(ctx, query, status, time.Now(), id)
	}
	if err != nil {
		return fmt.Errorf("failed to update customer status: %w", err)
	}
	return nil
}

func (r *CustomerRepository) UpdateBalance(ctx context.Context, id uuid.UUID, amount float64) error {
	query := `UPDATE customers SET balance = balance + $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, amount, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update customer ledger balance: %w", err)
	}
	return nil
}