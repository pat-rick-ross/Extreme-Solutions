package payment

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"Extreme-Solutions/internal/config"
	"Extreme-Solutions/internal/domain"
)

type PaystackService struct {
	cfg    *config.Config
	client *http.Client
}

func NewPaystackService(cfg *config.Config) *PaystackService {
	return &PaystackService{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *PaystackService) ProviderName() string {
	return "paystack"
}

func (p *PaystackService) InitiateSTKPush(ctx context.Context, phone string, amount float64, accountRef string) (string, error) {
	if p.client == nil {
		log.Panic("Paystack client is nil! Check NewPaystackService initialization.")
	}
	url := "https://api.paystack.co/charge"
	minorUnitsAmount := int(amount * 100)

	payload := map[string]interface{}{
		"email":  fmt.Sprintf("%s@extreme-isp.local", phone),
		"amount": minorUnitsAmount,
		"metadata": map[string]string{
			"account_reference": accountRef,
		},
		"mobile_money": map[string]string{
			"phone":    phone,
			"provider": "mpesa",
		},
	}

	bodyBytes, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}

	// Read cleanly from decoded config layout path
	req.Header.Set("Authorization", "Bearer "+p.cfg.Paystack.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("paystack charge request failed with status: %d", resp.StatusCode)
	}

	return fmt.Sprintf("pstk_ref_%d", time.Now().UnixNano()), nil
}

func (p *PaystackService) VerifyWebhookSignature(payload []byte, signatureHeader string) bool {
	hash := hmac.New(sha512.New, []byte(p.cfg.Paystack.SecretKey))
	hash.Write(payload)
	expectedSignature := hex.EncodeToString(hash.Sum(nil))
	return expectedSignature == signatureHeader
}

func (p *PaystackService) ParseWebhookEvent(payload []byte) (*domain.Payment, error) {
	type PaystackEvent struct {
		Event string `json:"event"`
		Data  struct {
			Reference string    `json:"reference"`
			Amount    float64   `json:"amount"`
			Status    string    `json:"status"`
			PaidAt    time.Time `json:"paid_at"`
			Channel   string    `json:"channel"`
			Customer  struct {
				Email string `json:"email"`
				Phone string `json:"phone"`
			} `json:"customer"`
		} `json:"data"`
	}

	var event PaystackEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("failed to decode paystack event structure: %w", err)
	}

	if event.Event != "charge.success" || event.Data.Status != "success" {
		return nil, errors.New("unhandled non-success event received")
	}

	return &domain.Payment{
		Reference:   event.Data.Reference,
		Amount:      event.Data.Amount / 100.0,
		Method:      event.Data.Channel,
		Provider:    "Paystack",
		Status:      "completed",
		MpesaPhone:  event.Data.Customer.Phone,
		Description: fmt.Sprintf("Fallback processing settlement via channel: %s", event.Data.Channel),
		Metadata:    payload,
		PaidAt:      &event.Data.PaidAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}
