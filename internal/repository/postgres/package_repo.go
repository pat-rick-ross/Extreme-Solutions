package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/isp-billing/internal/domain"
)

type packageRepository struct {
	db *pgxpool.Pool
}

func NewPackageRepository(db *pgxpool.Pool) *packageRepository {
	return &packageRepository{db: db}
}

func (r *packageRepository) Create(ctx context.Context, pkg *domain.InternetPackage) error {
	if pkg.ID == uuid.Nil {
		pkg.ID = uuid.New()
	}

	query := `
		INSERT INTO packages (id, name, description, speed, price, validity_days, is_active, bandwidth_up, bandwidth_down, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.Exec(ctx, query,
		pkg.ID,
		pkg.Name,
		pkg.Description,
		pkg.Speed,
		pkg.Price,
		pkg.ValidityDays,
		pkg.IsActive,
		pkg.BandwidthUp,
		pkg.BandwidthDown,
		pkg.CreatedAt,
		pkg.UpdatedAt,
	)

	return err
}

func (r *packageRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.InternetPackage, error) {
	query := `
		SELECT id, name, description, speed, price, validity_days, is_active, bandwidth_up, bandwidth_down, created_at, updated_at
		FROM packages
		WHERE id = $1
	`

	var pkg domain.InternetPackage
	err := r.db.QueryRow(ctx, query, id).Scan(
		&pkg.ID,
		&pkg.Name,
		&pkg.Description,
		&pkg.Speed,
		&pkg.Price,
		&pkg.ValidityDays,
		&pkg.IsActive,
		&pkg.BandwidthUp,
		&pkg.BandwidthDown,
		&pkg.CreatedAt,
		&pkg.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get package: %w", err)
	}

	return &pkg, nil
}

func (r *packageRepository) List(ctx context.Context) ([]*domain.InternetPackage, error) {
	query := `
		SELECT id, name, description, speed, price, validity_days, is_active, bandwidth_up, bandwidth_down, created_at, updated_at
		FROM packages
		ORDER BY price ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}
	defer rows.Close()

	var packages []*domain.InternetPackage
	for rows.Next() {
		var pkg domain.InternetPackage
		err := rows.Scan(
			&pkg.ID,
			&pkg.Name,
			&pkg.Description,
			&pkg.Speed,
			&pkg.Price,
			&pkg.ValidityDays,
			&pkg.IsActive,
			&pkg.BandwidthUp,
			&pkg.BandwidthDown,
			&pkg.CreatedAt,
			&pkg.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan package: %w", err)
		}
		packages = append(packages, &pkg)
	}

	return packages, nil
}

func (r *packageRepository) ListActive(ctx context.Context) ([]*domain.InternetPackage, error) {
	query := `
		SELECT id, name, description, speed, price, validity_days, is_active, bandwidth_up, bandwidth_down, created_at, updated_at
		FROM packages
		WHERE is_active = true
		ORDER BY price ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list active packages: %w", err)
	}
	defer rows.Close()

	var packages []*domain.InternetPackage
	for rows.Next() {
		var pkg domain.InternetPackage
		err := rows.Scan(
			&pkg.ID,
			&pkg.Name,
			&pkg.Description,
			&pkg.Speed,
			&pkg.Price,
			&pkg.ValidityDays,
			&pkg.IsActive,
			&pkg.BandwidthUp,
			&pkg.BandwidthDown,
			&pkg.CreatedAt,
			&pkg.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan package: %w", err)
		}
		packages = append(packages, &pkg)
	}

	return packages, nil
}

func (r *packageRepository) Update(ctx context.Context, pkg *domain.InternetPackage) error {
	query := `
		UPDATE packages
		SET name = $1, description = $2, speed = $3, price = $4, validity_days = $5, is_active = $6, bandwidth_up = $7, bandwidth_down = $8, updated_at = $9
		WHERE id = $10
	`

	_, err := r.db.Exec(ctx, query,
		pkg.Name,
		pkg.Description,
		pkg.Speed,
		pkg.Price,
		pkg.ValidityDays,
		pkg.IsActive,
		pkg.BandwidthUp,
		pkg.BandwidthDown,
		time.Now(),
		pkg.ID,
	)

	return err
}

func (r *packageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM packages WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
