package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"Extreme-Solutions/internal/domain"
	"Extreme-Solutions/internal/repository"
	"github.com/go-chi/chi/v5" // Ensure you have this for URL parameters
	"github.com/google/uuid"
)

type InvoiceHandler struct {
	invoiceRepo  repository.InvoiceRepository
	customerRepo repository.CustomerRepository
}

func NewInvoiceHandler(invoiceRepo repository.InvoiceRepository, customerRepo repository.CustomerRepository) *InvoiceHandler {
	return &InvoiceHandler{
		invoiceRepo:  invoiceRepo,
		customerRepo: customerRepo,
	}
}

func (h *InvoiceHandler) List(w http.ResponseWriter, r *http.Request) {
	// Since your repo lacks a GetAll, you might need to add one to your repository interface
	// For now, this assumes you add GetAll() to your InvoiceRepository
	invoices, err := h.invoiceRepo.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch invoices: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invoices)
}

func (h *InvoiceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	invoice, err := h.invoiceRepo.GetByID(r.Context(), id)
	if err != nil || invoice == nil {
		http.Error(w, "Invoice not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invoice)
}

func (h *InvoiceHandler) Generate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CustomerID string  `json:"customer_id"`
		Amount     float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	custID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		http.Error(w, "Invalid Customer ID", http.StatusBadRequest)
		return
	}

	newInvoice := &domain.Invoice{
		ID:         uuid.New(),
		CustomerID: custID,
		Amount:     req.Amount,
		Total:      req.Amount,
		Status:     "unpaid",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := h.invoiceRepo.Create(r.Context(), newInvoice); err != nil {
		http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newInvoice)
}

func (h *InvoiceHandler) GetPDF(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/pdf")
	w.WriteHeader(http.StatusOK)
}
