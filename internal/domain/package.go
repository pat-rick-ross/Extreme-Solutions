package domain

import (
	"time"

	"github.com/google/uuid"
)

type InternetPackage struct {
	ID            uuid.UUID `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	Description   string    `json:"description" db:"description"`
	Speed         string    `json:"speed" db:"speed"`
	Price         float64   `json:"price" db:"price"`
	ValidityDays  int       `json:"validity_days" db:"validity_days"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	BandwidthUp   int       `json:"bandwidth_up" db:"bandwidth_up"`   // in Kbps
	BandwidthDown int       `json:"bandwidth_down" db:"bandwidth_down"` // in Kbps
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type PackageCreateRequest struct {
	Name          string  `json:"name" validate:"required"`
	Description   string  `json:"description"`
	Speed         string  `json:"speed" validate:"required"`
	Price         float64 `json:"price" validate:"required,gt=0"`
	ValidityDays  int     `json:"validity_days" validate:"required,min=1"`
	BandwidthUp   int     `json:"bandwidth_up" validate:"required,min=64"`
	BandwidthDown int     `json:"bandwidth_down" validate:"required,min=128"`
}

type PackageUpdateRequest struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Speed         string  `json:"speed"`
	Price         float64 `json:"price" validate:"omitempty,gt=0"`
	ValidityDays  int     `json:"validity_days" validate:"omitempty,min=1"`
	BandwidthUp   int     `json:"bandwidth_up" validate:"omitempty,min=64"`
	BandwidthDown int     `json:"bandwidth_down" validate:"omitempty,min=128"`
	IsActive      *bool   `json:"is_active"`
}
