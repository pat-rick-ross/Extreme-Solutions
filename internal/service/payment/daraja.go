package payment

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"Extreme-Solutions/internal/config"
)

type DarajaService struct {
	cfg    *config.Config
	client *http.Client
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
}

type STKPushRequest struct {
	BusinessShortCode string `json:"BusinessShortCode"`
	Password          string `json:"Password"`
	Timestamp         string `json:"Timestamp"`
	TransactionType   string `json:"TransactionType"`
	Amount            int    `json:"Amount"`
	PartyA            string `json:"PartyA"`
	PartyB            string `json:"PartyB"`
	PhoneNumber       string `json:"PhoneNumber"`
	CallBackURL       string `json:"CallBackURL"`
	AccountReference  string `json:"AccountReference"`
	TransactionDesc   string `json:"TransactionDesc"`
}

type STKPushResponse struct {
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	CustomerMessage     string `json:"CustomerMessage"`
}

func NewDarajaService(cfg *config.Config) *DarajaService {
	return &DarajaService{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (d *DarajaService) ProviderName() string {
	return "daraja"
}

func (d *DarajaService) getBaseURL() string {
	if d.cfg.Mpesa.Environment == "production" {
		return "https://api.safaricom.co.ke"
	}
	return "https://sandbox.safaricom.co.ke"
}

func (d *DarajaService) generateToken(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/oauth/v1/generate?grant_type=client_credentials", d.getBaseURL())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	auth := fmt.Sprintf("%s:%s", d.cfg.Mpesa.ConsumerKey, d.cfg.Mpesa.ConsumerSecret)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))

	resp, err := d.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("daraja token auth returned status: %d", resp.StatusCode)
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", err
	}

	return authResp.AccessToken, nil
}

func (d *DarajaService) InitiateSTKPush(ctx context.Context, phone string, amount float64, accountRef string) (string, error) {
	token, err := d.generateToken(ctx)
	if err != nil {
		return "", fmt.Errorf("daraja token generation failed: %w", err)
	}

	timestamp := time.Now().Format("20060102150405")
	rawPassword := fmt.Sprintf("%s%s%s", d.cfg.Mpesa.Shortcode, d.cfg.Mpesa.Passkey, timestamp)
	password := base64.StdEncoding.EncodeToString([]byte(rawPassword))

	payload := STKPushRequest{
		BusinessShortCode: d.cfg.Mpesa.Shortcode,
		Password:          password,
		Timestamp:         timestamp,
		TransactionType:   "CustomerPayBillOnline",
		Amount:            int(amount),
		PartyA:            phone,
		PartyB:            d.cfg.Mpesa.Shortcode,
		PhoneNumber:       phone,
		CallBackURL:       d.cfg.Mpesa.CallbackURL,
		AccountReference:  accountRef,
		TransactionDesc:   "ISP Subscription Renewal",
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/mpesa/stkpush/v1/processrequest", d.getBaseURL())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var pushResp STKPushResponse
	if err := json.NewDecoder(resp.Body).Decode(&pushResp); err != nil {
		return "", err
	}

	if pushResp.ResponseCode != "0" {
		return "", fmt.Errorf("daraja rejected push request: %s", pushResp.ResponseDescription)
	}

	return pushResp.CheckoutRequestID, nil
}
