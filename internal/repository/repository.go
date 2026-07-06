package repository

import (
	"context"
	"extreme-solutions/internal/domain"

	"github.com/google/uuid"
)

type CustomerRepository interface {
	Create(ctx context.Context, customer *domain.Customer) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error)
	GetByPhone(ctx context.Context, phone string) (*domain.Customer, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateBalance(ctx context.Context, id uuid.UUID, amount float64) error
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
}

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	GetByReference(ctx context.Context, reference string) (*domain.Payment, error)
}