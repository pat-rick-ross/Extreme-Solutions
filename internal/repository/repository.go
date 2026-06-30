package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/isp-billing/internal/domain"
)

type CustomerRepository interface {
	Create(ctx context.Context, customer *domain.Customer) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error)
	GetByEmail(ctx context.Context, email string) (*domain.Customer, error)
	GetByPhone(ctx context.Context, phone string) (*domain.Customer, error)
	Update(ctx context.Context, customer *domain.Customer) error
	UpdateBalance(ctx context.Context, id uuid.UUID, amount float64) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.CustomerStatus) error
	List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*domain.Customer, int64, error)
	ListActive(ctx context.Context) ([]*domain.Customer, error)
	ListSuspended(ctx context.Context) ([]*domain.Customer, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type InvoiceRepository interface {
	Create(ctx context.Context, invoice *domain.Invoice) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Invoice, error)
	GetByNumber(ctx context.Context, number string) (*domain.Invoice, error)
	GetByCustomerID(ctx context.Context, customerID uuid.UUID, page, pageSize int) ([]*domain.Invoice, int64, error)
	List(ctx context.Context, filter *domain.InvoiceFilter) ([]*domain.Invoice, int64, error)
	ListPending(ctx context.Context) ([]*domain.Invoice, error)
	ListOverdue(ctx context.Context) ([]*domain.Invoice, error)
	Update(ctx context.Context, invoice *domain.Invoice) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.InvoiceStatus) error
	MarkAsPaid(ctx context.Context, id uuid.UUID, paidAt time.Time) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
	GetByReference(ctx context.Context, reference string) (*domain.Payment, error)
	GetByInvoiceID(ctx context.Context, invoiceID uuid.UUID) (*domain.Payment, error)
	GetByMpesaReceipt(ctx context.Context, receipt string) (*domain.Payment, error)
	ListByCustomerID(ctx context.Context, customerID uuid.UUID, page, pageSize int) ([]*domain.Payment, int64, error)
	ListPending(ctx context.Context) ([]*domain.Payment, error)
	Update(ctx context.Context, payment *domain.Payment) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus) error
	CompletePayment(ctx context.Context, id uuid.UUID, receipt string) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type PackageRepository interface {
	Create(ctx context.Context, pkg *domain.InternetPackage) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.InternetPackage, error)
	List(ctx context.Context) ([]*domain.InternetPackage, error)
	ListActive(ctx context.Context) ([]*domain.InternetPackage, error)
	Update(ctx context.Context, pkg *domain.InternetPackage) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type NASRepository interface {
	Create(ctx context.Context, nas *domain.NASDevice) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.NASDevice, error)
	List(ctx context.Context) ([]*domain.NASDevice, error)
	Update(ctx context.Context, nas *domain.NASDevice) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type MikroTikUserRepository interface {
	Create(ctx context.Context, user *domain.MikroTikUser) error
	GetByID(ctx context.Context, id string) (*domain.MikroTikUser, error)
	GetByCustomerID(ctx context.Context, customerID uuid.UUID) (*domain.MikroTikUser, error)
	GetByUsername(ctx context.Context, username string) (*domain.MikroTikUser, error)
	ListByNAS(ctx context.Context, nasID uuid.UUID) ([]*domain.MikroTikUser, error)
	Update(ctx context.Context, user *domain.MikroTikUser) error
	Delete(ctx context.Context, id string) error
}
