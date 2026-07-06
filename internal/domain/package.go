package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Package struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Speed         string    `json:"speed"`
	Price         float64   `json:"price"`
	ValidityDays  int       `json:"validity_days"`
	IsActive      bool      `json:"is_active"`
	BandwidthUp   int       `json:"bandwidth_up"`   // in Kbps
	BandwidthDown int       `json:"bandwidth_down"` // in Kbps
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (p *Package) Validate() error {
	if p.Name == "" {
		return errors.New("package name is required")
	}
	if p.Price <= 0 {
		return errors.New("package price must be greater than zero")
	}
	if p.BandwidthUp <= 0 || p.BandwidthDown <= 0 {
		return errors.New("bandwidth limits must be greater than zero")
	}
	return nil
}