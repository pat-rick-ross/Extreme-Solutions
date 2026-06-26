package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/isp-billing/internal/domain"
)

type customerRepository struct {
	db *pgxpool.Pool
}

func NewCustomerRepository(db *pgxpool.Pool) *customerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) Create(ctx context.Context, customer *domain.Customer) error {
	if customer.ID == uuid.Nil {
		customer.ID = uuid.New()
	}

	query := `
		INSERT INTO customers (id, first_name, last_name, email, phone, address, package_id, balance, status, mikrotik_id, mikrotik_user, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	now := time.Now()
	customer.CreatedAt = now
	customer.UpdatedAt = now
	customer.Status = domain.StatusPending
	customer.Balance = 0

	_, err := r.db.Exec(ctx, query,
		customer.ID,
		customer.FirstName,
		customer.LastName,
		customer.Email,
		customer.Phone,
		customer.Address,
		customer.PackageID,
		customer.Balance,
		customer.Status,
		customer.MikroTikID,
		customer.MikroTikUser,
		customer.CreatedAt,
		customer.UpdatedAt,
	)

	return err
}

func (r *customerRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, address, package_id, balance, status, mikrotik_id, mikrotik_user, created_at, updated_at, suspended_at, last_login_at
		FROM customers
		WHERE id = $1
	`

	var customer domain.Customer
	err := r.db.QueryRow(ctx, query, id).Scan(
		&customer.ID,
		&customer.FirstName,
		&customer.LastName,
		&customer.Email,
		&customer.Phone,
		&customer.Address,
		&customer.PackageID,
		&customer.Balance,
		&customer.Status,
		&customer.MikroTikID,
		&customer.MikroTikUser,
		&customer.CreatedAt,
		&customer.UpdatedAt,
		&customer.SuspendedAt,
		&customer.LastLoginAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &customer, nil
}

func (r *customerRepository) GetByEmail(ctx context.Context, email string) (*domain.Customer, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, address, package_id, balance, status, mikrotik_id, mikrotik_user, created_at, updated_at, suspended_at, last_login_at
		FROM customers
		WHERE email = $1
	`

	var customer domain.Customer
	err := r.db.QueryRow(ctx, query, email).Scan(
		&customer.ID,
		&customer.FirstName,
		&customer.LastName,
		&customer.Email,
		&customer.Phone,
		&customer.Address,
		&customer.PackageID,
		&customer.Balance,
		&customer.Status,
		&customer.MikroTikID,
		&customer.MikroTikUser,
		&customer.CreatedAt,
		&customer.UpdatedAt,
		&customer.SuspendedAt,
		&customer.LastLoginAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get customer by email: %w", err)
	}

	return &customer, nil
}

func (r *customerRepository) GetByPhone(ctx context.Context, phone string) (*domain.Customer, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, address, package_id, balance, status, mikrotik_id, mikrotik_user, created_at, updated_at, suspended_at, last_login_at
		FROM customers
		WHERE phone = $1
	`

	var customer domain.Customer
	err := r.db.QueryRow(ctx, query, phone).Scan(
		&customer.ID,
		&customer.FirstName,
		&customer.LastName,
		&customer.Email,
		&customer.Phone,
		&customer.Address,
		&customer.PackageID,
		&customer.Balance,
		&customer.Status,
		&customer.MikroTikID,
		&customer.MikroTikUser,
		&customer.CreatedAt,
		&customer.UpdatedAt,
		&customer.SuspendedAt,
		&customer.LastLoginAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get customer by phone: %w", err)
	}

	return &customer, nil
}

func (r *customerRepository) Update(ctx context.Context, customer *domain.Customer) error {
	query := `
		UPDATE customers
		SET first_name = $1, last_name = $2, email = $3, phone = $4, address = $5, package_id = $6, balance = $7, status = $8, updated_at = $9
		WHERE id = $10
	`

	customer.UpdatedAt = time.Now()

	_, err := r.db.Exec(ctx, query,
		customer.FirstName,
		customer.LastName,
		customer.Email,
		customer.Phone,
		customer.Address,
		customer.PackageID,
		customer.Balance,
		customer.Status,
		customer.UpdatedAt,
		customer.ID,
	)

	return err
}

func (r *customerRepository) UpdateBalance(ctx context.Context, id uuid.UUID, amount float64) error {
	query := `
		UPDATE customers
		SET balance = balance + $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, amount, time.Now(), id)
	return err
}

func (r *customerRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.CustomerStatus) error {
	query := `
		UPDATE customers
		SET status = $1, suspended_at = $2, updated_at = $3
		WHERE id = $4
	`

	var suspendedAt *time.Time
	if status == domain.StatusSuspended {
		now := time.Now()
		suspendedAt = &now
	}

	_, err := r.db.Exec(ctx, query, status, suspendedAt, time.Now(), id)
	return err
}

func (r *customerRepository) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*domain.Customer, int64, error) {
	var conditions []string
	var args []interface{}
	argCount := 1

	if status, ok := filter["status"]; ok {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argCount))
		args = append(args, status)
		argCount++
	}

	if packageID, ok := filter["package_id"]; ok {
		conditions = append(conditions, fmt.Sprintf("package_id = $%d", argCount))
		args = append(args, packageID)
		argCount++
	}

	if search, ok := filter["search"]; ok {
		conditions = append(conditions, fmt.Sprintf("(first_name ILIKE $%d OR last_name ILIKE $%d OR email ILIKE $%d OR phone ILIKE $%d)", argCount, argCount, argCount, argCount))
		searchTerm := "%" + search.(string) + "%"
		args = append(args, searchTerm, searchTerm, searchTerm, searchTerm)
		argCount += 4
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM customers %s", whereClause)
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count customers: %w", err)
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`
		SELECT id, first_name, last_name, email, phone, address, package_id, balance, status, mikrotik_id, mikrotik_user, created_at, updated_at, suspended_at, last_login_at
		FROM customers
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	args = append(args, pageSize, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list customers: %w", err)
	}
	defer rows.Close()

	var customers []*domain.Customer
	for rows.Next() {
		var customer domain.Customer
		err := rows.Scan(
			&customer.ID,
			&customer.FirstName,
			&customer.LastName,
			&customer.Email,
			&customer.Phone,
			&customer.Address,
			&customer.PackageID,
			&customer.Balance,
			&customer.Status,
			&customer.MikroTikID,
			&customer.MikroTikUser,
			&customer.CreatedAt,
			&customer.UpdatedAt,
			&customer.SuspendedAt,
			&customer.LastLoginAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, &customer)
	}

	return customers, total, nil
}

func (r *customerRepository) ListActive(ctx context.Context) ([]*domain.Customer, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, address, package_id, balance, status, mikrotik_id, mikrotik_user, created_at, updated_at, suspended_at, last_login_at
		FROM customers
		WHERE status = $1
	`

	rows, err := r.db.Query(ctx, query, domain.StatusActive)
	if err != nil {
		return nil, fmt.Errorf("failed to list active customers: %w", err)
	}
	defer rows.Close()

	var customers []*domain.Customer
	for rows.Next() {
		var customer domain.Customer
		err := rows.Scan(
			&customer.ID,
			&customer.FirstName,
			&customer.LastName,
			&customer.Email,
			&customer.Phone,
			&customer.Address,
			&customer.PackageID,
			&customer.Balance,
			&customer.Status,
			&customer.MikroTikID,
			&customer.MikroTikUser,
			&customer.CreatedAt,
			&customer.UpdatedAt,
			&customer.SuspendedAt,
			&customer.LastLoginAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, &customer)
	}

	return customers, nil
}

func (r *customerRepository) ListSuspended(ctx context.Context) ([]*domain.Customer, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, address, package_id, balance, status, mikrotik_id, mikrotik_user, created_at, updated_at, suspended_at, last_login_at
		FROM customers
		WHERE status = $1
	`

	rows, err := r.db.Query(ctx, query, domain.StatusSuspended)
	if err != nil {
		return nil, fmt.Errorf("failed to list suspended customers: %w", err)
	}
	defer rows.Close()

	var customers []*domain.Customer
	for rows.Next() {
		var customer domain.Customer
		err := rows.Scan(
			&customer.ID,
			&customer.FirstName,
			&customer.LastName,
			&customer.Email,
			&customer.Phone,
			&customer.Address,
			&customer.PackageID,
			&customer.Balance,
			&customer.Status,
			&customer.MikroTikID,
			&customer.MikroTikUser,
			&customer.CreatedAt,
			&customer.UpdatedAt,
			&customer.SuspendedAt,
			&customer.LastLoginAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, &customer)
	}

	return customers, nil
}

func (r *customerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM customers WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
