package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"Extreme-Solutions/internal/domain"

	"github.com/google/uuid"
)

type PackageRepository struct {
	db *sql.DB
}

// NewPackageRepository initializes the postgres implementation for ISP tiers
func NewPackageRepository(db *sql.DB) *PackageRepository {
	return &PackageRepository{db: db}
}

func (r *PackageRepository) GetAllActive(ctx context.Context) ([]domain.Package, error) {
	query := `
		SELECT id, name, description, speed, price, validity_days, is_active, bandwidth_up, bandwidth_down, created_at, updated_at
		FROM packages
		WHERE is_active = true
		ORDER BY price ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active packages: %w", err)
	}
	defer rows.Close()

	var pkgs []domain.Package
	for rows.Next() {
		var p domain.Package
		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Speed, &p.Price,
			&p.ValidityDays, &p.IsActive, &p.BandwidthUp, &p.BandwidthDown,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan package row: %w", err)
		}
		pkgs = append(pkgs, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return pkgs, nil
}

func (r *PackageRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Package, error) {
	query := `
		SELECT id, name, description, speed, price, validity_days, is_active, bandwidth_up, bandwidth_down, created_at, updated_at
		FROM packages
		WHERE id = $1
	`
	var p domain.Package
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Speed, &p.Price,
		&p.ValidityDays, &p.IsActive, &p.BandwidthUp, &p.BandwidthDown,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package by id: %w", err)
	}
	return &p, nil
}
