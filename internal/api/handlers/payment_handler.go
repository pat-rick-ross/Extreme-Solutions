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
}

func NewPaymentHandler(
	paymentRepo repository.PaymentRepository,
	invoiceRepo repository.InvoiceRepository,
	customerRepo repository.CustomerRepository,
	darajaSvc *payment.DarajaService,
	paystackSvc *payment.PaystackService,
) *PaymentHandler {
	return &PaymentHandler{
		paymentRepo:  paymentRepo,
		invoiceRepo:  invoiceRepo,
		customerRepo: customerRepo,
		darajaSvc:    darajaSvc,
		paystackSvc:  paystackSvc,
	}
}

func (h *PaymentHandler) InitiatePayment(w http.ResponseWriter, r *http.Request) {
	var req domain.PaymentInitiateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Native input field validation fallback (Aligned with domain properties)
	if req.InvoiceID == "" || req.Amount <= 0 || req.Phone == "" {
		respondError(w, http.StatusBadRequest, "Missing required payment parameters (invoice_id, amount, phone)")
		return
	}

	// Get invoice
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

	// Get customer
	customer, err := h.customerRepo.GetByID(r.Context(), invoice.CustomerID)
	if err != nil {
		log.Printf("[ERROR] Failed to resolve client metadata profile for customer %s: %v", invoice.CustomerID, err)
		respondError(w, http.StatusInternalServerError, "Failed to initiate payment")
		return
	}

	// Create dynamic transaction record context layout
	paymentRecord := &domain.Payment{
		InvoiceID:  &invoice.ID, // Assigned using memory pointer location allocation
		CustomerID: customer.ID,
		Amount:     req.Amount,
		Status:     "pending",
		Method:     "mpesa",
		Provider:   "mpesa",
		MpesaPhone: req.Phone,
		Reference:  uuid.New().String(),
	}

	if err := h.paymentRepo.Create(r.Context(), paymentRecord); err != nil {
		log.Printf("[ERROR] Failed to record pending billing entry ledger map: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to initiate payment")
		return
	}

	// Dynamic fallback for the Daraja statement account reference identifier
	accountRef := invoice.Reference
	if accountRef == "" {
		accountRef = "INV-" + paymentRecord.Reference[:8]
	}

	// Match signature: want (context.Context, string, float64, string)
	stkRef, err := h.darajaSvc.InitiateSTKPush(r.Context(), req.Phone, req.Amount, accountRef)
	if err != nil {
		log.Printf("[ERROR] Safaricom Daraja STK execution gateway exception encountered: %v", err)

		// Fallback mutation strategy to handle the absence of UpdateStatus interface methods
		paymentRecord.Status = "failed"
		_ = h.paymentRepo.Update(r.Context(), paymentRecord)

		respondError(w, http.StatusInternalServerError, "Failed to broadcast payment request prompt to device")
		return
	}

	// Safe serialization for raw jsonb database payload slices targeting tracking fields
	metadataMap := map[string]string{
		"daraja_tracking_reference": stkRef,
	}
	if metadataBytes, err := json.Marshal(metadataMap); err == nil {
		paymentRecord.Metadata = metadataBytes
		if err := h.paymentRepo.Update(r.Context(), paymentRecord); err != nil {
			log.Printf("[WARN] Error persisting transaction reference identifier metadata: %v", err)
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"payment":          paymentRecord,
		"tracking_id":      stkRef,
		"customer_message": "STK Push prompt broadcasted successfully to your mobile phone.",
	})
}

func (h *PaymentHandler) GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	// 1. Grab the raw string param directly from Chi's multiplexer
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Missing payment ID")
		return
	}

	// 2. Pass the raw string "id" directly since your repo expects a string value
	paymentRecord, err := h.paymentRepo.GetByID(r.Context(), id)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch target payment record details %s: %v", id, err)
		respondError(w, http.StatusInternalServerError, "Failed to fetch transaction record metadata")
		return
	}
	if paymentRecord == nil {
		respondError(w, http.StatusNotFound, "Transaction data file profile not found")
		return
	}

	respondJSON(w, http.StatusOK, paymentRecord)
}

func (h *PaymentHandler) GetCustomerPayments(w http.ResponseWriter, r *http.Request) {
	customerID, ok := r.Context().Value("customer_id").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Invalid context context session scope validation mapping")
		return
	}

	page := getIntParam(r, "page", 1)
	pageSize := getIntParam(r, "page_size", 20)

	payments, total, err := h.paymentRepo.ListByCustomerID(r.Context(), customerID, page, pageSize)
	if err != nil {
		log.Printf("[ERROR] Failed to query dynamic subscriber list collection for owner account ID %s: %v", customerID, err)
		respondError(w, http.StatusInternalServerError, "Failed to parse records data trace")
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
