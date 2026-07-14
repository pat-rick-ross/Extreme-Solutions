package domain

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Invoice struct {
	ID          uuid.UUID      `json:"id"`
	CustomerID  uuid.UUID      `json:"customer_id"`
	Number      string         `json:"number"`
	Amount      float64        `json:"amount"`
	Tax         float64        `json:"tax"`
	Total       float64        `json:"total"`
	Status      string         `json:"status"` // "pending", "unpaid", "paid", "cancelled"
	Description sql.NullString `json:"description,omitempty"`
	PeriodStart time.Time      `json:"period_start"`
	PeriodEnd   time.Time      `json:"period_end"`
	DueDate     time.Time      `json:"due_date"`
	PaidAt      *time.Time     `json:"paid_at,omitempty"`
	PDFURL      sql.NullString `json:"pdf_url,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Reference   string         `json:"reference"`
}
