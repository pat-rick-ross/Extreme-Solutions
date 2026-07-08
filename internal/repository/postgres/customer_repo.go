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

func (r *CustomerRepository) Update(ctx context.Context, c *domain.Customer) error {
	query := `
		UPDATE customers 
		SET first_name = $1, last_name = $2, email = $3, phone = $4, address = $5, 
		    package_id = $6, balance = $7, status = $8, updated_at = $9, 
		    suspended_at = $10, last_login_at = $11, mikrotik_user = $12, mikrotik_id = $13
		WHERE id = $14
	`
	_, err := r.db.ExecContext(ctx, query,
		c.FirstName,
		c.LastName,
		c.Email,
		c.Phone,
		c.Address,
		c.PackageID,
		c.Balance,
		c.Status,
		time.Now(),
		c.SuspendedAt,
		c.LastLoginAt,
		c.MikrotikUser,
		c.MikrotikID,
		c.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update customer record parameters: %w", err)
	}
	return nil
}

func (r *CustomerRepository) GetByEmail(ctx context.Context, email string) (*domain.Customer, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, address, package_id, balance, status, created_at, updated_at, suspended_at, last_login_at
		FROM customers WHERE email = $1
	`
	var c domain.Customer
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&c.ID, &c.FirstName, &c.LastName, &c.Email, &c.Phone, &c.Address,
		&c.PackageID, &c.Balance, &c.Status, &c.CreatedAt, &c.UpdatedAt,
		&c.SuspendedAt, &c.LastLoginAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch customer by email: %w", err)
	}
	return &c, nil
}

// Add this method to internal/repository/postgres/customer_repo.go

func (r *CustomerRepository) List(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*domain.Customer, int64, error) {
	// Base query string construction block
	query := `
		SELECT id, first_name, last_name, email, phone, address, package_id, balance, status, created_at, updated_at, suspended_at, last_login_at, COUNT(*) OVER()
		FROM customers
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute query customer list: %w", err)
	}
	defer rows.Close()

	var customers []*domain.Customer
	var totalCount int64 = 0

	for rows.Next() {
		var c domain.Customer
		err := rows.Scan(
			&c.ID, &c.FirstName, &c.LastName, &c.Email, &c.Phone, &c.Address,
			&c.PackageID, &c.Balance, &c.Status, &c.CreatedAt, &c.UpdatedAt,
			&c.SuspendedAt, &c.LastLoginAt, &totalCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan customer dataset row: %w", err)
		}
		customers = append(customers, &c)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	return customers, totalCount, nil
}

// Add this method to internal/repository/postgres/customer_repo.go

func (r *CustomerRepository) ListActive(ctx context.Context) ([]*domain.Customer, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, address, package_id, balance, status, created_at, updated_at, suspended_at, last_login_at
		FROM customers
		WHERE status = 'active'
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active customer list: %w", err)
	}
	defer rows.Close()

	var customers []*domain.Customer
	for rows.Next() {
		var c domain.Customer
		err := rows.Scan(
			&c.ID, &c.FirstName, &c.LastName, &c.Email, &c.Phone, &c.Address,
			&c.PackageID, &c.Balance, &c.Status, &c.CreatedAt, &c.UpdatedAt,
			&c.SuspendedAt, &c.LastLoginAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan active customer row: %w", err)
		}
		customers = append(customers, &c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return customers, nil
}

// Add this method at the bottom of your customer_repo.go file

// Delete safely purges a customer profile block from the database ledger.
func (r *CustomerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM customers WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to purge customer record: %w", err)
	}
	return nil
}
