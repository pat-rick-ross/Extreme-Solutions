package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"Extreme-Solutions/internal/domain"
	"Extreme-Solutions/internal/service/payment" // Points to your core service package
)

type WebhookHandler struct {
	paystackService *payment.PaystackService
	darajaService   *payment.DarajaService
	reconciler      *payment.PaymentReconciler // Core service layer worker dependency
}

func NewWebhookHandler(p *payment.PaystackService, d *payment.DarajaService, r *payment.PaymentReconciler) *WebhookHandler {
	return &WebhookHandler{
		paystackService: p,
		darajaService:   d,
		reconciler:      r,
	}
}

// 1. DARAJA WEBHOOK
type DarajaCallbackPayload struct {
	Body struct {
		StkCallback struct {
			MerchantRequestID string `json:"MerchantRequestID"`
			CheckoutRequestID string `json:"CheckoutRequestID"`
			ResultCode        int    `json:"ResultCode"`
			ResultDesc        string `json:"ResultDesc"`
			CallbackMetadata  struct {
				Item []struct {
					Name  string      `json:"Name"`
					Value interface{} `json:"Value,omitempty"`
				} `json:"Item"`
			} `json:"CallbackMetadata"`
		} `json:"stkCallback"`
	} `json:"Body"`
}

func (h *WebhookHandler) HandleDarajaWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload DarajaCallbackPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	cb := payload.Body.StkCallback
	if cb.ResultCode != 0 {
		log.Printf("[Daraja] Callback transaction cancelled/failed: %s", cb.ResultDesc)
		w.WriteHeader(http.StatusOK)
		return
	}

	var amount float64
	var mpesaReceipt string
	var phoneNumber string

	for _, item := range cb.CallbackMetadata.Item {
		switch item.Name {
		case "Amount":
			if val, ok := item.Value.(float64); ok {
				amount = val
			}
		case "MpesaReceiptNumber":
			if val, ok := item.Value.(string); ok {
				mpesaReceipt = val
			}
		case "PhoneNumber":
			if val, ok := item.Value.(string); ok {
				phoneNumber = val
			}
		}
	}

	paymentRecord := &domain.Payment{
		Reference:    mpesaReceipt,
		MpesaReceipt: mpesaReceipt,
		MpesaPhone:   phoneNumber,
		Amount:       amount,
		Method:       "mpesa",
		Provider:     "Safaricom_Daraja",
		Status:       "completed",
		Description:  "Primary Daraja API Paybill confirmation hit.",
	}

	log.Printf("[✓ SUCCESS] Daraja confirmed payment of %.2f from %s", paymentRecord.Amount, paymentRecord.MpesaPhone)

	accountRef := "INV-XXXX" // Extract this variable value dynamically from custom attributes or metadata payloads

	// Fixed compilation block syntax issue by initializing tracking error variables using ':=' safely
	if err := h.reconciler.ProcessSuccessfulPayment(r.Context(), paymentRecord, accountRef); err != nil {
		log.Printf("[ERROR] Webhook processing error encountered: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ResultCode": 0, "ResultDesc": "Accepted"}`))
}

// 2. PAYSTACK WEBHOOK
func (h *WebhookHandler) HandlePaystackWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	payloadBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	signature := r.Header.Get("x-paystack-signature")
	if !h.paystackService.VerifyWebhookSignature(payloadBytes, signature) {
		log.Printf("[SECURITY WARNING] Invalid Paystack signature drop encountered.")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	paymentRecord, err := h.paystackService.ParseWebhookEvent(payloadBytes)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Printf("[✓ FALLBACK SUCCESS] Paystack processed transaction of %.2f from phone %s", paymentRecord.Amount, paymentRecord.MpesaPhone)

	accountRef := "INV-XXXX" // Extract this variable value dynamically from custom attributes or metadata payloads
	if err := h.reconciler.ProcessSuccessfulPayment(r.Context(), paymentRecord, accountRef); err != nil {
		log.Printf("[ERROR] Webhook processing error encountered: %v", err)
	}

	w.WriteHeader(http.StatusOK)
}
