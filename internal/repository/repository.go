package repository

import (
	"Extreme-Solutions/internal/domain"
	"context"
	"time"

	"github.com/google/uuid"
)

type CustomerRepository interface {
	Create(ctx context.Context, customer *domain.Customer) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error)
	GetByEmail(ctx context.Context, email string) (*domain.Customer, error)
	GetByPhone(ctx context.Context, phone string) (*domain.Customer, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*domain.Customer, int64, error)
	UpdateBalance(ctx context.Context, id uuid.UUID, amount float64) error
	ListActive(ctx context.Context) ([]*domain.Customer, error)
	Update(ctx context.Context, customer *domain.Customer) error
}

type CacheRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}

type PackageRepository interface {
	GetAllActive(ctx context.Context) ([]domain.Package, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Package, error)
}

type InvoiceRepository interface {
	Create(ctx context.Context, invoice *domain.Invoice) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Invoice, error)
	GetUnpaidByCustomerID(ctx context.Context, customerID uuid.UUID) ([]domain.Invoice, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, paidAt *string) error
	ExistsByReference(ctx context.Context, reference string) (bool, error)
	GetByReference(ctx context.Context, reference string) (*domain.Invoice, error)
	Update(ctx context.Context, invoice *domain.Invoice) error
	MarkAsPaid(ctx context.Context, id uuid.UUID, paidAt time.Time) error // <-- Make sure this is here
	ListOverdue(ctx context.Context, asOf time.Time) ([]*domain.Invoice, error)
	GetAll(ctx context.Context) ([]*domain.Invoice, error) // <--- Add this line
}

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	Update(ctx context.Context, payment *domain.Payment) error
	ListByCustomerID(ctx context.Context, customerID uuid.UUID, page, pageSize int) ([]*domain.Payment, int64, error)
	GetByID(ctx context.Context, reference string) (*domain.Payment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error       // <-- Make sure this is here
	ListPending(ctx context.Context) ([]*domain.Payment, error)                // <-- Make sure this is here
	CompletePayment(ctx context.Context, id uuid.UUID, reference string) error // <-- Make sure this is here
}
