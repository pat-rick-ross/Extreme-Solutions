package payment

import (
	"Extreme-Solutions/internal/config"
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
)

type IntaSendService struct {
	cfg    *config.Config
	client *resty.Client
}

func NewIntaSendService(cfg *config.Config) *IntaSendService {
	return &IntaSendService{
		cfg:    cfg,
		client: resty.New().SetAuthToken(cfg.IntaSend.SecretKey),
	}
}

func (i *IntaSendService) ProviderName() string { return "intasend" }

func (i *IntaSendService) InitiateSTKPush(ctx context.Context, phone string, amount float64, accountRef string) (string, error) {
	resp, err := i.client.R().
		SetContext(ctx).
		SetBody(map[string]interface{}{
			"public_key":   i.cfg.IntaSend.PublishableKey,
			"amount":       amount,
			"currency":     "KES",
			"phone_number": phone,
			"account":      accountRef,
		}).
		Post("https://sandbox.intasend.com/api/v1/payment/mpesa-stk-push/")

	if err != nil {
		return "", fmt.Errorf("intasend request failed: %w", err)
	}

	if resp.IsError() {
		return "", fmt.Errorf("intasend api error: %s", resp.String())
	}

	return "success", nil
}
