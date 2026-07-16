package network

import (
	"context"
	"fmt"
	"log"
	"time"

	"Extreme-Solutions/internal/config"
	"Extreme-Solutions/internal/domain"
	"github.com/go-routeros/routeros"
)

type Provisioner struct {
	client *routeros.Client
	cfg    config.MikrotikConfig // Casing aligned with your global config struct configuration properties
}

func NewProvisioner(cfg config.MikrotikConfig) *Provisioner {
	return &Provisioner{
		cfg: cfg,
	}
}

// Connect instantiates an active connection pool over the RouterOS API socket port
func (p *Provisioner) Connect() error {
	client, err := routeros.Dial(
		fmt.Sprintf("%s:%d", p.cfg.Host, p.cfg.Port),
		p.cfg.Username,
		p.cfg.Password,
	)
	if err != nil {
		return fmt.Errorf("failed to connect to MikroTik router: %w", err)
	}
	p.client = client
	return nil
}

// Close gracefully terminates the active socket connection to prevent resource leaks
func (p *Provisioner) Close() {
	if p.client != nil {
		p.client.Close()
		p.client = nil
	}
}

func (p *Provisioner) ProvisionPPPoE(ctx context.Context, customer *domain.Customer, pkg *domain.Package) error {
	if p.client == nil {
		if err := p.Connect(); err != nil {
			return err
		}
	}

	// Generate clear deterministic structural access credentials
	username := fmt.Sprintf("ISP-%s", customer.ID.String()[:8])
	password := generatePassword()

	// Execute command via proper RouterOS structural arguments array layout
	_, err := p.client.Run(
		"/ppp/secret/add",
		"=name="+username,
		"=password="+password,
		"=service=pppoe",
		"=profile=default",
		"=disabled=no",
	)
	if err != nil {
		return fmt.Errorf("failed to write customer PPPoE secret: %w", err)
	}

	// Attach operational queue control rules instantly if subscription profile exists
	if pkg != nil {
		// Calculate limits based on package values
		limitAttr := fmt.Sprintf("%dk/%dk", int(pkg.Price)*1024, int(pkg.Price)*1024)
		_, err = p.client.Run(
			"/queue/simple/add",
			"=name="+username,
			"=target="+username, // Maps queue targeting parameter to PPPoE user handle interface
			"=max-limit="+limitAttr,
			"=comment=Managed automatically via Extreme Solutions core engine",
		)
		if err != nil {
			log.Printf("[ERROR] Failed to auto-bind profile bandwidth limits for customer %s: %v\n", customer.ID.String(), err)
		}
	}

	// Synchronize back directly into the lowercased domain fields
	customer.MikrotikUser = username
	customer.MikrotikID = username

	return nil
}

func (p *Provisioner) SuspendUser(ctx context.Context, customer *domain.Customer) error {
	if p.client == nil {
		if err := p.Connect(); err != nil {
			return err
		}
	}

	// Using numbers= instead of .id= allows targeting named items by their unique key name safely
	_, err := p.client.Run(
		"/ppp/secret/disable",
		"=numbers="+customer.MikrotikUser,
	)
	if err != nil {
		return fmt.Errorf("failed to toggle restriction bit for target subscriber user: %w", err)
	}

	// Terminate any active sessions to force the router to drop their connection instantly
	_, _ = p.client.Run(
		"/ppp/active/remove",
		"=numbers="+customer.MikrotikUser,
	)

	return nil
}

func (p *Provisioner) ReactivateUser(ctx context.Context, customer *domain.Customer) error {
	if p.client == nil {
		if err := p.Connect(); err != nil {
			return err
		}
	}

	_, err := p.client.Run(
		"/ppp/secret/enable",
		"=numbers="+customer.MikrotikUser,
	)
	if err != nil {
		return fmt.Errorf("failed to lift suspension constraint rule flags: %w", err)
	}

	return nil
}

func (p *Provisioner) SetBandwidth(ctx context.Context, customer *domain.Customer, up, down int) error {
	if p.client == nil {
		if err := p.Connect(); err != nil {
			return err
		}
	}

	limitAttr := fmt.Sprintf("%dk/%dk", up*1024, down*1024)
	_, err := p.client.Run(
		"/queue/simple/set",
		"=numbers="+customer.MikrotikUser,
		"=max-limit="+limitAttr,
	)
	return err
}

func generatePassword() string {
	return fmt.Sprintf("ISP%d", time.Now().UnixNano()%1000000)
}
