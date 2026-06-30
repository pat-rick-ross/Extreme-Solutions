package domain

import (
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string
type PaymentMethod string
type PaymentProvider string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

const (
	MethodMPESA PaymentMethod = "mpesa"
	MethodCash  PaymentMethod = "cash"
	MethodBank  PaymentMethod = "bank"
)

const (
	ProviderMPESA PaymentProvider = "mpesa"
)

type Payment struct {
	ID          uuid.UUID        `json:"id" db:"id"`
	InvoiceID   uuid.UUID        `json:"invoice_id" db:"invoice_id"`
	Invoice     *Invoice         `json:"invoice,omitempty"`
	CustomerID  uuid.UUID        `json:"customer_id" db:"customer_id"`
	Amount      float64          `json:"amount" db:"amount"`
	Status      PaymentStatus    `json:"status" db:"status"`
	Method      PaymentMethod    `json:"method" db:"method"`
	Provider    PaymentProvider  `json:"provider" db:"provider"`
	Reference   string           `json:"reference" db:"reference"`
	MpesaReceipt string          `json:"mpesa_receipt,omitempty" db:"mpesa_receipt"`
	MpesaPhone  string           `json:"mpesa_phone,omitempty" db:"mpesa_phone"`
	Description string           `json:"description" db:"description"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	CompletedAt *time.Time       `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at" db:"updated_at"`
}

type PaymentInitiateRequest struct {
	InvoiceID   string  `json:"invoice_id" validate:"required"`
	PhoneNumber string  `json:"phone_number" validate:"required,phone"`
	Amount      float64 `json:"amount" validate:"required,gt=0"`
}

type MPESAWebhookPayload struct {
	Body struct {
		StkCallback struct {
			MerchantRequestID string `json:"MerchantRequestID"`
			CheckoutRequestID string `json:"CheckoutRequestID"`
			ResultCode        int    `json:"ResultCode"`
			ResultDesc        string `json:"ResultDesc"`
			CallbackMetadata  struct {
				Item []struct {
					Name  string      `json:"Name"`
					Value interface{} `json:"Value"`
				} `json:"Item"`
			} `json:"CallbackMetadata"`
		} `json:"stkCallback"`
	} `json:"Body"`
}
