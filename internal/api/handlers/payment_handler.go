package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"Extreme-Solutions/internal/domain"
	"Extreme-Solutions/internal/repository"
	"Extreme-Solutions/internal/service/payment"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PaymentHandler struct {
	paymentRepo  repository.PaymentRepository
	invoiceRepo  repository.InvoiceRepository
	customerRepo repository.CustomerRepository
	darajaSvc    *payment.DarajaService
	paystackSvc  *payment.PaystackService
	intasendSvc  *payment.IntaSendService // <--- THIS IS THE MISSING FIELD
}

func NewPaymentHandler(
	paymentRepo repository.PaymentRepository,
	invoiceRepo repository.InvoiceRepository,
	customerRepo repository.CustomerRepository,
	darajaSvc *payment.DarajaService,
	paystackSvc *payment.PaystackService,
	intasendSvc *payment.IntaSendService, // Add this parameter

) *PaymentHandler {
	return &PaymentHandler{
		paymentRepo:  paymentRepo,
		invoiceRepo:  invoiceRepo,
		customerRepo: customerRepo,
		darajaSvc:    darajaSvc,
		paystackSvc:  paystackSvc,
		intasendSvc:  intasendSvc, // Add this assignment
	}
}

func (h *PaymentHandler) InitiatePayment(w http.ResponseWriter, r *http.Request) {
	var req domain.PaymentInitiateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields including the new Gateway selection
	if req.InvoiceID == "" || req.Amount <= 0 || req.Phone == "" || req.Gateway == "" {
		respondError(w, http.StatusBadRequest, "Missing required payment parameters (invoice_id, amount, phone, gateway)")
		return
	}

	// Dynamically select the gateway based on request
	var gateway payment.PaymentGateway
	switch req.Gateway {
	case "daraja":
		gateway = h.darajaSvc
	case "intasend":
		gateway = h.intasendSvc
	case "paystack":
		gateway = h.paystackSvc
	default:
		respondError(w, http.StatusBadRequest, "Unsupported payment gateway: "+req.Gateway)
		return
	}

	invoiceID, err := uuid.Parse(req.InvoiceID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid invoice ID format")
		return
	}

	invoice, err := h.invoiceRepo.GetByID(r.Context(), invoiceID)
	if err != nil {
		log.Printf("[ERROR] Failed to query invoice %s: %v", invoiceID, err)
		respondError(w, http.StatusInternalServerError, "Failed to initiate payment")
		return
	}
	if invoice == nil {
		respondError(w, http.StatusNotFound, "Invoice not found")
		return
	}

	if invoice.Status == "paid" {
		respondError(w, http.StatusBadRequest, "Invoice has already been paid")
		return
	}

	customer, err := h.customerRepo.GetByID(r.Context(), invoice.CustomerID)
	if err != nil {
		log.Printf("[ERROR] Failed to resolve client metadata profile for customer %s: %v", invoice.CustomerID, err)
		respondError(w, http.StatusInternalServerError, "Failed to initiate payment")
		return
	}

	paymentRecord := &domain.Payment{
		InvoiceID:  &invoice.ID,
		CustomerID: customer.ID,
		Amount:     req.Amount,
		Status:     "pending",
		Method:     req.Gateway,
		Provider:   gateway.ProviderName(),
		MpesaPhone: req.Phone,
		Reference:  uuid.New().String(),
	}

	if err := h.paymentRepo.Create(r.Context(), paymentRecord); err != nil {
		log.Printf("[ERROR] Failed to record pending billing entry ledger map: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to initiate payment")
		return
	}

	accountRef := invoice.Reference
	if accountRef == "" {
		accountRef = "INV-" + paymentRecord.Reference[:8]
	}

	// Polymorphic call to the selected gateway
	stkRef, err := gateway.InitiateSTKPush(r.Context(), req.Phone, req.Amount, accountRef)
	if err != nil {
		log.Printf("[ERROR] Gateway %s execution exception encountered: %v", req.Gateway, err)

		paymentRecord.Status = "failed"
		_ = h.paymentRepo.Update(r.Context(), paymentRecord)

		respondError(w, http.StatusInternalServerError, "Failed to broadcast payment request prompt to device")
		return
	}

	metadataMap := map[string]string{
		"tracking_reference": stkRef,
		"gateway":            req.Gateway,
	}
	if metadataBytes, err := json.Marshal(metadataMap); err == nil {
		paymentRecord.Metadata = metadataBytes
		_ = h.paymentRepo.Update(r.Context(), paymentRecord)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"payment":          paymentRecord,
		"tracking_id":      stkRef,
		"customer_message": "Payment prompt broadcasted successfully via " + req.Gateway,
	})
}

func (h *PaymentHandler) GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Missing payment ID")
		return
	}

	paymentRecord, err := h.paymentRepo.GetByID(r.Context(), id)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch target payment record details %s: %v", id, err)
		respondError(w, http.StatusInternalServerError, "Failed to fetch transaction record metadata")
		return
	}
	if paymentRecord == nil {
		respondError(w, http.StatusNotFound, "Transaction data profile not found")
		return
	}

	respondJSON(w, http.StatusOK, paymentRecord)
}

func (h *PaymentHandler) GetCustomerPayments(w http.ResponseWriter, r *http.Request) {
	customerID, ok := r.Context().Value("customer_id").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Invalid context session scope")
		return
	}

	page := getIntParam(r, "page", 1)
	pageSize := getIntParam(r, "page_size", 20)

	payments, total, err := h.paymentRepo.ListByCustomerID(r.Context(), customerID, page, pageSize)
	if err != nil {
		log.Printf("[ERROR] Failed to query subscriber payment list: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to parse records")
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
