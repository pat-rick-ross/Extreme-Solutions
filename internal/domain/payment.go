package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID           uuid.UUID       `json:"id"`
	InvoiceID    *uuid.UUID      `json:"invoice_id,omitempty"`
	CustomerID   uuid.UUID       `json:"customer_id"`
	Amount       float64         `json:"amount"`
	Status       string          `json:"status"` // "pending", "completed", "failed"
	Method       string          `json:"method"` // "mpesa", "bank", "cash"
	Provider     string          `json:"provider"`
	Reference    string          `json:"reference"` // M-Pesa Code (e.g. SBR123XYZ)
	MpesaReceipt string          `json:"mpesa_receipt,omitempty"`
	MpesaPhone   string          `json:"mpesa_phone,omitempty"`
	Description  string          `json:"description,omitempty"`
	Metadata     json.RawMessage `json:"metadata,omitempty"` // Maps directly to Postgres JSONB type
	PaidAt       *time.Time      `json:"completed_at,omitempty"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type PaymentInitiateRequest struct {
	Phone     string  `json:"phone"`
	Amount    float64 `json:"amount"`
	InvoiceID string  `json:"invoice_id"`
}
