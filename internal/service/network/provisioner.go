package network

import (
	"context"
	"fmt"
	"time"

	"github.com/go-routeros/routeros"
	"github.com/your-org/isp-billing/internal/config"
	"github.com/your-org/isp-billing/internal/domain"
	"github.com/your-org/isp-billing/internal/pkg/logger"
)

type Provisioner struct {
	client *routeros.Client
	cfg    config.MikroTikConfig
}

func NewProvisioner(cfg config.MikroTikConfig) *Provisioner {
	return &Provisioner{
		cfg: cfg,
	}
}

func (p *Provisioner) Connect() error {
	client, err := routeros.Dial(
		fmt.Sprintf("%s:%d", p.cfg.Host, p.cfg.Port),
		p.cfg.Username,
		p.cfg.Password,
	)
	if err != nil {
		return fmt.Errorf("failed to connect to MikroTik: %w", err)
	}
	p.client = client
	return nil
}

func (p *Provisioner) ProvisionPPPoE(ctx context.Context, customer *domain.Customer, pkg *domain.InternetPackage) error {
	if p.client == nil {
		if err := p.Connect(); err != nil {
			return err
		}
	}

	// Generate username and password
	username := fmt.Sprintf("ISP-%s", customer.ID.String()[:8])
	password := generatePassword()

	// Create PPPoE secret
	cmd := fmt.Sprintf(
		"/ppp/secret/add name=%s password=%s service=pppoe profile=default disabled=no",
		username, password,
	)

	_, err := p.client.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to create PPPoE secret: %w", err)
	}

	// Set bandwidth limits if needed
	if pkg != nil {
		cmd = fmt.Sprintf(
			"/queue/simple/add name=%s target-addresses=0.0.0.0/0 dst-address=0.0.0.0/0 limit-at=%d/%d max-limit=%d/%d",
			username,
			pkg.BandwidthUp*1024, pkg.BandwidthDown*1024,
			pkg.BandwidthUp*1024, pkg.BandwidthDown*1024,
		)
		_, err := p.client.Run(cmd)
		if err != nil {
			logger.Error("Failed to set bandwidth limits", map[string]interface{}{
				"customer": customer.ID,
				"error":    err,
			})
		}
	}

	// Update customer with MikroTik credentials
	customer.MikroTikUser = username
	customer.MikroTikID = username // In practice, you'd get the ID from response

	return nil
}

func (p *Provisioner) SuspendUser(ctx context.Context, customer *domain.Customer) error {
	if p.client == nil {
		if err := p.Connect(); err != nil {
			return err
		}
	}

	cmd := fmt.Sprintf("/ppp/secret/disable where name=%s", customer.MikroTikUser)
	_, err := p.client.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to suspend user: %w", err)
	}

	return nil
}

func (p *Provisioner) ReactivateUser(ctx context.Context, customer *domain.Customer) error {
	if p.client == nil {
		if err := p.Connect(); err != nil {
			return err
		}
	}

	cmd := fmt.Sprintf("/ppp/secret/enable where name=%s", customer.MikroTikUser)
	_, err := p.client.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to reactivate user: %w", err)
	}

	return nil
}

func (p *Provisioner) SetBandwidth(ctx context.Context, customer *domain.Customer, up, down int) error {
	if p.client == nil {
		if err := p.Connect(); err != nil {
			return err
		}
	}

	cmd := fmt.Sprintf(
		"/queue/simple/set %s limit-at=%d/%d max-limit=%d/%d",
		customer.MikroTikUser,
		up*1024, down*1024,
		up*1024, down*1024,
	)
	_, err := p.client.Run(cmd)
	return err
}

func generatePassword() string {
	return fmt.Sprintf("ISP%d", time.Now().UnixNano()%1000000)
}
