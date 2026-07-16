package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"Extreme-Solutions/internal/domain"
	"Extreme-Solutions/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CustomerHandler struct {
	customerRepo repository.CustomerRepository
	packageRepo  repository.PackageRepository
}

func NewCustomerHandler(customerRepo repository.CustomerRepository, packageRepo repository.PackageRepository) *CustomerHandler {
	return &CustomerHandler{
		customerRepo: customerRepo,
		packageRepo:  packageRepo,
	}
}

func (h *CustomerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CustomerCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.FirstName == "" || req.LastName == "" || req.Email == "" || req.Phone == "" || req.PackageID == "" {
		respondError(w, http.StatusBadRequest, "Missing required profile registration parameters")
		return
	}

	existing, err := h.customerRepo.GetByEmail(r.Context(), req.Email)
	if err != nil {
		log.Printf("[ERROR] Failed to check existing customer by email: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create customer")
		return
	}
	if existing != nil {
		respondError(w, http.StatusConflict, "Email already registered")
		return
	}

	existing, err = h.customerRepo.GetByPhone(r.Context(), req.Phone)
	if err != nil {
		log.Printf("[ERROR] Failed to check existing customer by phone: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create customer")
		return
	}
	if existing != nil {
		respondError(w, http.StatusConflict, "Phone already registered")
		return
	}

	packageID, err := uuid.Parse(req.PackageID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid package ID")
		return
	}

	pkg, err := h.packageRepo.GetByID(r.Context(), packageID)
	if err != nil {
		log.Printf("[ERROR] Failed to get package %s: %v", packageID, err)
		respondError(w, http.StatusInternalServerError, "Failed to create customer")
		return
	}
	if pkg == nil {
		respondError(w, http.StatusNotFound, "Package not found")
		return
	}

	customer := &domain.Customer{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Phone:     req.Phone,
		// Using the helper to convert string to sql.NullString
		Address:   NewNullString(req.Address),
		PackageID: &packageID,
	}

	if err := h.customerRepo.Create(r.Context(), customer); err != nil {
		log.Printf("[ERROR] Failed to persist new customer schema: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create customer")
		return
	}

	respondJSON(w, http.StatusCreated, customer)
}

func (h *CustomerHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	customerID, err := uuid.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	customer, err := h.customerRepo.GetByID(r.Context(), customerID)
	if err != nil {
		log.Printf("[ERROR] Failed to get customer %s: %v", customerID, err)
		respondError(w, http.StatusInternalServerError, "Failed to get customer")
		return
	}

	if customer == nil {
		respondError(w, http.StatusNotFound, "Customer not found")
		return
	}

	respondJSON(w, http.StatusOK, customer)
}

func (h *CustomerHandler) List(w http.ResponseWriter, r *http.Request) {
	page := getIntParam(r, "page", 1)
	pageSize := getIntParam(r, "page_size", 20)

	filter := make(map[string]interface{})
	if status := r.URL.Query().Get("status"); status != "" {
		filter["status"] = status
	}
	if packageID := r.URL.Query().Get("package_id"); packageID != "" {
		filter["package_id"] = packageID
	}
	if search := r.URL.Query().Get("search"); search != "" {
		filter["search"] = search
	}

	customers, total, err := h.customerRepo.List(r.Context(), filter, page, pageSize)
	if err != nil {
		log.Printf("[ERROR] Failed to query paginated customer pool: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list customers")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":        customers,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

func (h *CustomerHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	customerID, err := uuid.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	var req domain.CustomerUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	customer, err := h.customerRepo.GetByID(r.Context(), customerID)
	if err != nil {
		log.Printf("[ERROR] Failed to locate customer reference %s before update: %v", customerID, err)
		respondError(w, http.StatusInternalServerError, "Failed to update customer")
		return
	}
	if customer == nil {
		respondError(w, http.StatusNotFound, "Customer not found")
		return
	}

	if req.FirstName != "" {
		customer.FirstName = req.FirstName
	}
	if req.LastName != "" {
		customer.LastName = req.LastName
	}
	if req.Email != "" {
		customer.Email = req.Email
	}
	if req.Phone != "" {
		customer.Phone = req.Phone
	}
	// Using the helper here as well
	if req.Address != "" {
		customer.Address = NewNullString(req.Address)
	}
	if req.PackageID != "" {
		packageID, err := uuid.Parse(req.PackageID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid package ID")
			return
		}
		customer.PackageID = &packageID
	}

	if err := h.customerRepo.Update(r.Context(), customer); err != nil {
		log.Printf("[ERROR] Failed to execute update statement transaction: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update customer")
		return
	}

	respondJSON(w, http.StatusOK, customer)
}

func (h *CustomerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	customerID, err := uuid.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	if err := h.customerRepo.Delete(r.Context(), customerID); err != nil {
		log.Printf("[ERROR] Failed to isolate delete hook for target customer %s: %v", customerID, err)
		respondError(w, http.StatusInternalServerError, "Failed to delete customer")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
