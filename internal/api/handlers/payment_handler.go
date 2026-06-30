package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/your-org/isp-billing/internal/domain"
	"github.com/your-org/isp-billing/internal/pkg/logger"
	"github.com/your-org/isp-billing/internal/pkg/validator"
	"github.com/your-org/isp-billing/internal/repository"
	"github.com/your-org/isp-billing/internal/service/payment"
)

type PaymentHandler struct {
	paymentRepo  repository.PaymentRepository
	invoiceRepo  repository.InvoiceRepository
	customerRepo repository.CustomerRepository
	mpesaService *payment.MPESAService
	validator    *validator.Validator
}

func NewPaymentHandler(
	paymentRepo repository.PaymentRepository,
	invoiceRepo repository.InvoiceRepository,
	customerRepo repository.CustomerRepository,
	mpesaService *payment.MPESAService,
) *PaymentHandler {
	return &PaymentHandler{
		paymentRepo:  paymentRepo,
		invoiceRepo:  invoiceRepo,
		customerRepo: customerRepo,
		mpesaService: mpesaService,
		validator:    validator.New(),
	}
}

func (h *PaymentHandler) InitiatePayment(w http.ResponseWriter, r *http.Request) {
	var req domain.PaymentInitiateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Validate(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get invoice
	invoiceID, err := uuid.Parse(req.InvoiceID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid invoice ID")
		return
	}

	invoice, err := h.invoiceRepo.GetByID(r.Context(), invoiceID)
	if err != nil {
		logger.Error("Failed to get invoice", map[string]interface{}{"error": err})
		respondError(w, http.StatusInternalServerError, "Failed to initiate payment")
		return
	}
	if invoice == nil {
		respondError(w, http.StatusNotFound, "Invoice not found")
		return
	}

	if invoice.Status == domain.InvoiceStatusPaid {
		respondError(w, http.StatusBadRequest, "Invoice already paid")
		return
	}

	// Get customer
	customer, err := h.customerRepo.GetByID(r.Context(), invoice.CustomerID)
	if err != nil {
		logger.Error("Failed to get customer", map[string]interface{}{"error": err})
		respondError(w, http.StatusInternalServerError, "Failed to initiate payment")
		return
	}

	// Create payment record
	payment := &domain.Payment{
		InvoiceID:  invoice.ID,
		CustomerID: customer.ID,
		Amount:     req.Amount,
		Status:     domain.PaymentStatusPending,
		Method:     domain.MethodMPESA,
		Provider:   domain.ProviderMPESA,
		MpesaPhone: req.PhoneNumber,
		Reference:  uuid.New().String(),
	}

	if err := h.paymentRepo.Create(r.Context(), payment); err != nil {
		logger.Error("Failed to create payment", map[string]interface{}{"error": err})
		respondError(w, http.StatusInternalServerError, "Failed to initiate payment")
		return
	}

	// Initiate STK push
	stkResp, err := h.mpesaService.InitiateSTKPush(r.Context(), &req)
	if err != nil {
		logger.Error("Failed to initiate STK push", map[string]interface{}{"error": err})
		// Update payment status to failed
		_ = h.paymentRepo.UpdateStatus(r.Context(), payment.ID, domain.PaymentStatusFailed)
		respondError(w, http.StatusInternalServerError, "Failed to initiate payment")
		return
	}

	// Update payment with checkout ID
	payment.Metadata = map[string]interface{}{
		"checkout_request_id": stkResp.CheckoutRequestID,
		"merchant_request_id": stkResp.MerchantRequestID,
	}
	if err := h.paymentRepo.Update(r.Context(), payment); err != nil {
		logger.Warn("Failed to update payment metadata", map[string]interface{}{"error": err})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"payment":            payment,
		"checkout_request_id": stkResp.CheckoutRequestID,
		"customer_message":   stkResp.CustomerMessage,
	})
}

func (h *PaymentHandler) GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	paymentID, err := uuid.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid payment ID")
		return
	}

	payment, err := h.paymentRepo.GetByID(r.Context(), paymentID)
	if err != nil {
		logger.Error("Failed to get payment", map[string]interface{}{"error": err})
		respondError(w, http.StatusInternalServerError, "Failed to get payment")
		return
	}
	if payment == nil {
		respondError(w, http.StatusNotFound, "Payment not found")
		return
	}

	respondJSON(w, http.StatusOK, payment)
}

func (h *PaymentHandler) GetCustomerPayments(w http.ResponseWriter, r *http.Request) {
	customerID := r.Context().Value("customer_id").(uuid.UUID)
	page := getIntParam(r, "page", 1)
	pageSize := getIntParam(r, "page_size", 20)

	payments, total, err := h.paymentRepo.ListByCustomerID(r.Context(), customerID, page, pageSize)
	if err != nil {
		logger.Error("Failed to list payments", map[string]interface{}{"error": err})
		respondError(w, http.StatusInternalServerError, "Failed to list payments")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":        payments,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}
