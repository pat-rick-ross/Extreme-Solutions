package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/your-org/isp-billing/internal/config"
	"github.com/your-org/isp-billing/internal/pkg/logger"
)

type SMSService struct {
	cfg        config.SMSConfig
	httpClient *http.Client
}

type AfricaTalkingRequest struct {
	Username string `json:"username"`
	To       string `json:"to"`
	Message  string `json:"message"`
	From     string `json:"from"`
}

type AfricaTalkingResponse struct {
	SMSMessageData struct {
		Recipients []struct {
			Number    string `json:"number"`
			Cost      string `json:"cost"`
			Status    string `json:"status"`
			MessageID string `json:"messageId"`
		} `json:"recipients"`
	} `json:"SMSMessageData"`
}

func NewSMSService(cfg config.SMSConfig) *SMSService {
	return &SMSService{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *SMSService) SendSMS(to, message string) error {
	if !s.cfg.Enabled {
		logger.Info("SMS service disabled, skipping send", map[string]interface{}{
			"to":      to,
			"message": message,
		})
		return nil
	}

	switch s.cfg.Provider {
	case "africastalking":
		return s.sendAfricaTalking(to, message)
	default:
		return fmt.Errorf("unsupported SMS provider: %s", s.cfg.Provider)
	}
}

func (s *SMSService) sendAfricaTalking(to, message string) error {
	url := "https://api.africastalking.com/version1/messaging"

	reqBody := AfricaTalkingRequest{
		Username: "sandbox", // Use your actual username in production
		To:       to,
		Message:  message,
		From:     s.cfg.SenderID,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal SMS request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create SMS request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apiKey", s.cfg.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SMS API returned status: %d", resp.StatusCode)
	}

	var africaResp AfricaTalkingResponse
	if err := json.NewDecoder(resp.Body).Decode(&africaResp); err != nil {
		return fmt.Errorf("failed to parse SMS response: %w", err)
	}

	if len(africaResp.SMSMessageData.Recipients) > 0 {
		recipient := africaResp.SMSMessageData.Recipients[0]
		if recipient.Status != "Success" {
			return fmt.Errorf("SMS sending failed with status: %s", recipient.Status)
		}
	}

	logger.Info("SMS sent successfully", map[string]interface{}{
		"to":      to,
		"message": message,
	})

	return nil
}

// Template methods
func (s *SMSService) SendInvoiceReminder(phone, customerName, invoiceNumber string, amount float64) error {
	message := fmt.Sprintf(
		"Dear %s, your invoice %s of KES %.2f is due. Please pay via M-PESA Paybill to avoid suspension.",
		customerName, invoiceNumber, amount,
	)
	return s.SendSMS(phone, message)
}

func (s *SMSService) SendPaymentConfirmation(phone, customerName, receipt string, amount float64) error {
	message := fmt.Sprintf(
		"Dear %s, payment of KES %.2f received. Receipt: %s. Thank you for choosing our service.",
		customerName, amount, receipt,
	)
	return s.SendSMS(phone, message)
}

func (s *SMSService) SendSuspensionWarning(phone, customerName string, daysOverdue int) error {
	message := fmt.Sprintf(
		"Dear %s, your account is %d days overdue. Service will be suspended in 24 hours. Pay now to avoid interruption.",
		customerName, daysOverdue,
	)
	return s.SendSMS(phone, message)
}

func (s *SMSService) SendSuspensionNotice(phone, customerName string) error {
	message := fmt.Sprintf(
		"Dear %s, your service has been suspended due to non-payment. Please pay to reactivate.",
		customerName,
	)
	return s.SendSMS(phone, message)
}

func (s *SMSService) SendReactivationNotice(phone, customerName string) error {
	message := fmt.Sprintf(
		"Dear %s, your service has been reactivated. Welcome back!",
		customerName,
	)
	return s.SendSMS(phone, message)
}
