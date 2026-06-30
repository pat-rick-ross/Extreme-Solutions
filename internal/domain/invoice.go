package domain

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	InvoiceStatusPending  InvoiceStatus = "pending"
	InvoiceStatusPaid     InvoiceStatus = "paid"
	InvoiceStatusOverdue  InvoiceStatus = "overdue"
	InvoiceStatusCanceled InvoiceStatus = "canceled"
)

type Invoice struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	CustomerID  uuid.UUID      `json:"customer_id" db:"customer_id"`
	Customer    *Customer      `json:"customer,omitempty"`
	Number      string         `json:"number" db:"number"`
	Amount      float64        `json:"amount" db:"amount"`
	Tax         float64        `json:"tax" db:"tax"`
	Total       float64        `json:"total" db:"total"`
	Status      InvoiceStatus  `json:"status" db:"status"`
	Description string         `json:"description" db:"description"`
	PeriodStart time.Time      `json:"period_start" db:"period_start"`
	PeriodEnd   time.Time      `json:"period_end" db:"period_end"`
	DueDate     time.Time      `json:"due_date" db:"due_date"`
	PaidAt      *time.Time     `json:"paid_at,omitempty" db:"paid_at"`
	PDFURL      string         `json:"pdf_url,omitempty" db:"pdf_url"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

type InvoiceCreateRequest struct {
	CustomerID  string    `json:"customer_id" validate:"required"`
	Amount      float64   `json:"amount" validate:"required,gt=0"`
	Description string    `json:"description"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	DueDate     time.Time `json:"due_date"`
}

type InvoiceFilter struct {
	CustomerID string
	Status     InvoiceStatus
	StartDate  time.Time
	EndDate    time.Time
	Page       int
	PageSize   int
}
