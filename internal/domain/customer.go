package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID           uuid.UUID  `json:"id"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	Address      string     `json:"address,omitempty"`
	PackageID    *uuid.UUID `json:"package_id,omitempty"` // Pointer handles nullable foreign keys cleanly
	Balance      float64    `json:"balance"`
	Status       string     `json:"status"` // "pending", "active", "suspended", "terminated"
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	SuspendedAt  *time.Time `json:"suspended_at,omitempty"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}

func (c *Customer) Validate() error {
	if c.FirstName == "" || c.LastName == "" {
		return errors.New("customer names cannot be empty")
	}
	if !strings.HasPrefix(c.Phone, "+254") && !strings.HasPrefix(c.Phone, "07") && !strings.HasPrefix(c.Phone, "01") {
		return errors.New("invalid Kenyan phone number format")
	}
	return nil
}