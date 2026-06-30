package domain

import (
	"time"

	"github.com/google/uuid"
)

type CustomerStatus string

const (
	StatusActive    CustomerStatus = "active"
	StatusSuspended CustomerStatus = "suspended"
	StatusPending   CustomerStatus = "pending"
	StatusInactive  CustomerStatus = "inactive"
)

type Customer struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	FirstName   string         `json:"first_name" db:"first_name"`
	LastName    string         `json:"last_name" db:"last_name"`
	Email       string         `json:"email" db:"email"`
	Phone       string         `json:"phone" db:"phone"`
	Address     string         `json:"address" db:"address"`
	PackageID   uuid.UUID      `json:"package_id" db:"package_id"`
	Package     *InternetPackage `json:"package,omitempty"`
	Balance     float64        `json:"balance" db:"balance"`
	Status      CustomerStatus `json:"status" db:"status"`
	MikroTikID  string         `json:"mikrotik_id,omitempty" db:"mikrotik_id"`
	MikroTikUser string        `json:"mikrotik_user,omitempty" db:"mikrotik_user"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
	SuspendedAt *time.Time     `json:"suspended_at,omitempty" db:"suspended_at"`
	LastLoginAt *time.Time     `json:"last_login_at,omitempty" db:"last_login_at"`
}

type CustomerCreateRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2"`
	LastName  string `json:"last_name" validate:"required,min=2"`
	Email     string `json:"email" validate:"required,email"`
	Phone     string `json:"phone" validate:"required,phone"`
	Address   string `json:"address"`
	PackageID string `json:"package_id" validate:"required"`
}

type CustomerUpdateRequest struct {
	FirstName string         `json:"first_name"`
	LastName  string         `json:"last_name"`
	Email     string         `json:"email" validate:"omitempty,email"`
	Phone     string         `json:"phone" validate:"omitempty,phone"`
	Address   string         `json:"address"`
	PackageID string         `json:"package_id"`
	Status    CustomerStatus `json:"status"`
}

type CustomerLoginResponse struct {
	Customer   *Customer `json:"customer"`
	AccessToken string  `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
