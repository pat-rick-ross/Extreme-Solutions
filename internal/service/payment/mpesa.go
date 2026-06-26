package payment

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/your-org/isp-billing/internal/config"
	"github.com/your-org/isp-billing/internal/domain"
	"github.com/your-org/isp-billing/internal/pkg/logger"
)

type MPESAService struct {
	cfg         config.MPESAConfig
	httpClient  *http.Client
	authToken   string
	tokenExpiry time.Time
}

// STKPushRequest represents the M-PESA STK Push request
type STKPushRequest struct {
	BusinessShortCode string `json:"BusinessShortCode"`
	Password          string `json:"Password"`
	Timestamp         string `json:"Timestamp"`
	TransactionType   string `json:"TransactionType"`
	Amount            string `json:"Amount"`
	PartyA            string `json:"PartyA"`
	PartyB            string `json:"PartyB"`
	PhoneNumber       string `json:"PhoneNumber"`
	CallBackURL       string `json:"CallBackURL"`
	AccountReference  string `json:"AccountReference"`
	TransactionDesc   string `json:"TransactionDesc"`
}

// STKPushResponse represents the M-PESA STK Push response
type STKPushResponse struct {
	MerchantRequestID string `json:"MerchantRequestID"`
	CheckoutRequestID string `json:"CheckoutRequestID"`
	ResponseCode      string `json:"ResponseCode"`
	ResponseDesc      string `json:"ResponseDesc"`
	CustomerMessage   string `json:"CustomerMessage"`
}

// AuthResponse represents the M-PESA OAuth response
type AuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func NewMPESAService(cfg config.MPESAConfig) *MPESAService {
	return &MPESAService{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// InitiateSTKPush initiates a STK Push transaction
func (m *MPESAService) InitiateSTKPush(ctx context.Context, req *domain.PaymentInitiateRequest) (*STKPushResponse, error) {
	// Validate configuration
	if m.cfg.ConsumerKey == "" || m.cfg.ConsumerSecret == "" {
		return nil, fmt.Errorf("M-PESA credentials not configured")
	}

	if m.cfg.Passkey == "" {
		return nil, fmt.Errorf("M-PESA Passkey not configured - required for STK Push")
	}

	// Get auth token
	if err := m.ensureAuth(ctx); err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	// Generate timestamp and password
	timestamp := time.Now().Format("20060102150405")
	password := m.generatePassword(timestamp)

	logger.Info("Initiating M-PESA STK Push", map[string]interface{}{
		"amount":      req.Amount,
		"phone":       req.PhoneNumber,
		"invoice_id":  req.InvoiceID,
		"shortcode":   m.cfg.ShortCode,
		"environment": m.cfg.Environment,
	})

	stkReq := STKPushRequest{
		BusinessShortCode: m.cfg.ShortCode,
		Password:          password,
		Timestamp:         timestamp,
		TransactionType:   "CustomerPayBillOnline",
		Amount:            fmt.Sprintf("%.0f", req.Amount),
		PartyA:            req.PhoneNumber,
		PartyB:            m.cfg.ShortCode,
		PhoneNumber:       req.PhoneNumber,
		CallBackURL:       m.cfg.CallbackURL,
		AccountReference:  req.InvoiceID,
		TransactionDesc:   "ISP Bill Payment",
	}

	jsonData, err := json.Marshal(stkReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal STK push request: %w", err)
	}

	// Log request (without sensitive data)
	logger.Debug("STK Push Request", map[string]interface{}{
		"shortcode": stkReq.BusinessShortCode,
		"amount":    stkReq.Amount,
		"phone":     stkReq.PhoneNumber,
		"reference": stkReq.AccountReference,
	})

	url := m.getBaseURL() + "/mpesa/stkpush/v1/processrequest"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+m.authToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send STK push: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	logger.Debug("M-PESA Response", map[string]interface{}{
		"status_code": resp.StatusCode,
		"body":        string(body),
	})

	var stkResp STKPushResponse
	if err := json.Unmarshal(body, &stkResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if stkResp.ResponseCode != "0" {
		return nil, fmt.Errorf("STK push failed: %s (Code: %s)", stkResp.ResponseDesc, stkResp.ResponseCode)
	}

	logger.Info("STK Push initiated successfully", map[string]interface{}{
		"checkout_request_id": stkResp.CheckoutRequestID,
		"merchant_request_id": stkResp.MerchantRequestID,
	})

	return &stkResp, nil
}

// QueryStatus queries the status of a transaction
func (m *MPESAService) QueryStatus(ctx context.Context, checkoutRequestID string) (string, error) {
	// Get auth token
	if err := m.ensureAuth(ctx); err != nil {
		return "", fmt.Errorf("failed to get auth token: %w", err)
	}

	// Build query request
	queryReq := map[string]interface{}{
		"BusinessShortCode": m.cfg.ShortCode,
		"Password":          m.generatePassword(time.Now().Format("20060102150405")),
		"Timestamp":         time.Now().Format("20060102150405"),
		"CheckoutRequestID": checkoutRequestID,
	}

	jsonData, err := json.Marshal(queryReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal query request: %w", err)
	}

	url := m.getBaseURL() + "/mpesa/stkpushquery/v1/query"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create query request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+m.authToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to query status: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read query response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse query response: %w", err)
	}

	// Extract result code
	if resultCode, ok := result["ResultCode"]; ok {
		if resultCode.(string) == "0" {
			return "completed", nil
		} else if resultCode.(string) == "1" {
			return "failed", nil
		}
	}

	return "pending", nil
}

// ensureAuth ensures we have a valid auth token
func (m *MPESAService) ensureAuth(ctx context.Context) error {
	// Check if token is still valid
	if m.authToken != "" && time.Now().Before(m.tokenExpiry) {
		return nil
	}

	logger.Info("Getting M-PESA OAuth token", map[string]interface{}{
		"environment": m.cfg.Environment,
	})

	url := m.getBaseURL() + "/oauth/v1/generate?grant_type=client_credentials"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(m.cfg.ConsumerKey + ":" + m.cfg.ConsumerSecret))
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	if authResp.AccessToken == "" {
		return fmt.Errorf("no access token in response")
	}

	m.authToken = authResp.AccessToken
	m.tokenExpiry = time.Now().Add(55 * time.Minute)

	logger.Info("M-PESA OAuth token obtained successfully", map[string]interface{}{
		"expires_in": authResp.ExpiresIn,
	})

	return nil
}

// generatePassword generates the M-PESA password
func (m *MPESAService) generatePassword(timestamp string) string {
	str := m.cfg.ShortCode + m.cfg.Passkey + timestamp
	hash := sha256.Sum256([]byte(str))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// getBaseURL returns the appropriate M-PESA API base URL
func (m *MPESAService) getBaseURL() string {
	if m.cfg.Environment == "production" {
		return "https://api.safaricom.co.ke"
	}
	return "https://sandbox.safaricom.co.ke"
}

// ProcessWebhook processes incoming M-PESA webhook
func (m *MPESAService) ProcessWebhook(payload []byte) (*MPESAWebhookResponse, error) {
	var webhookResp MPESAWebhookResponse
	if err := json.Unmarshal(payload, &webhookResp); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	logger.Info("Processing M-PESA webhook", map[string]interface{}{
		"result_code": webhookResp.Body.StkCallback.ResultCode,
		"result_desc": webhookResp.Body.StkCallback.ResultDesc,
	})

	return &webhookResp, nil
}

// MPESAWebhookResponse represents the M-PESA webhook payload
type MPESAWebhookResponse struct {
	Body struct {
		StkCallback struct {
			MerchantRequestID string `json:"MerchantRequestID"`
			CheckoutRequestID string `json:"CheckoutRequestID"`
			ResultCode        int    `json:"ResultCode"`
			ResultDesc        string `json:"ResultDesc"`
			CallbackMetadata  struct {
				Item []struct {
					Name  string      `json:"Name"`
					Value interface{} `json:"Value"`
				} `json:"Item"`
			} `json:"CallbackMetadata"`
		} `json:"stkCallback"`
	} `json:"Body"`
}

// ExtractReceipt extracts the receipt number from webhook response
func (w *MPESAWebhookResponse) ExtractReceipt() string {
	for _, item := range w.Body.StkCallback.CallbackMetadata.Item {
		if item.Name == "MpesaReceiptNumber" {
			if receipt, ok := item.Value.(string); ok {
				return receipt
			}
		}
	}
	return ""
}

// ExtractAmount extracts the amount from webhook response
func (w *MPESAWebhookResponse) ExtractAmount() float64 {
	for _, item := range w.Body.StkCallback.CallbackMetadata.Item {
		if item.Name == "Amount" {
			if amount, ok := item.Value.(float64); ok {
				return amount
			}
		}
	}
	return 0
}

// IsSuccessful checks if the webhook indicates a successful transaction
func (w *MPESAWebhookResponse) IsSuccessful() bool {
	return w.Body.StkCallback.ResultCode == 0
}
