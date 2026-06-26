package domain

import (
	"time"

	"github.com/google/uuid"
)

type NASDevice struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Host        string    `json:"host" db:"host"`
	Port        int       `json:"port" db:"port"`
	Username    string    `json:"username" db:"username"`
	Password    string    `json:"password" db:"password"`
	Type        string    `json:"type" db:"type"` // mikrotik, pfsense, etc.
	Location    string    `json:"location" db:"location"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	LastPingAt  *time.Time `json:"last_ping_at,omitempty" db:"last_ping_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type MikroTikUser struct {
	ID         string    `json:"id" db:"id"`
	CustomerID uuid.UUID `json:"customer_id" db:"customer_id"`
	NASID      uuid.UUID `json:"nas_id" db:"nas_id"`
	Username   string    `json:"username" db:"username"`
	Password   string    `json:"password" db:"password"`
	Profile    string    `json:"profile" db:"profile"`
	Uptime     int       `json:"uptime" db:"uptime"` // seconds
	BytesIn    int64     `json:"bytes_in" db:"bytes_in"`
	BytesOut   int64     `json:"bytes_out" db:"bytes_out"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type NetworkProvisionRequest struct {
	CustomerID   uuid.UUID `json:"customer_id" validate:"required"`
	Username     string    `json:"username" validate:"required"`
	Password     string    `json:"password" validate:"required,min=8"`
	BandwidthUp  int       `json:"bandwidth_up"`
	BandwidthDown int      `json:"bandwidth_down"`
}
