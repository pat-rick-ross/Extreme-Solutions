package payment

import (
	"context"
	"fmt"
	"log"
	"time"
)

type PaymentGateway interface {
	InitiateSTKPush(ctx context.Context, phone string, amount float64, accountRef string) (string, error)
	ProviderName() string
}

type PaymentOrchestrator struct {
	primary  PaymentGateway
	fallback PaymentGateway
}

func NewPaymentOrchestrator(primary PaymentGateway, fallback PaymentGateway) *PaymentOrchestrator {
	return &PaymentOrchestrator{
		primary:  primary,
		fallback: fallback,
	}
}

func (o *PaymentOrchestrator) RouteDynamicSTK(ctx context.Context, phone string, amount float64, accountRef string) (string, error) {
	log.Printf("[Orchestrator] Dispatching to Primary: %s", o.primary.ProviderName())

	primaryCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	checkoutID, err := o.primary.InitiateSTKPush(primaryCtx, phone, amount, accountRef)
	if err == nil {
		log.Printf("[Orchestrator] Primary push dispatched successfully. Tracking ID: %s", checkoutID)
		return checkoutID, nil
	}

	// Failover routing condition
	log.Printf("[CRITICAL] Primary gateway failed: %v. Engaging fallback tracking lane instantly...", err)

	fallbackCtx, fallbackCancel := context.WithTimeout(ctx, 10*time.Second)
	defer fallbackCancel()

	fallbackID, fallbackErr := o.fallback.InitiateSTKPush(fallbackCtx, phone, amount, accountRef)
	if fallbackErr != nil {
		return "", fmt.Errorf("both primary and fallback channels are structurally unreachable: %w", fallbackErr)
	}

	log.Printf("[Orchestrator] Failover push succeeded via: %s", o.fallback.ProviderName())
	return fallbackID, nil
}
