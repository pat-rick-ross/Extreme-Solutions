package notification

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"

	"github.com/your-org/isp-billing/internal/config"
	"github.com/your-org/isp-billing/internal/pkg/logger"
)

type EmailService struct {
	cfg config.EmailConfig
}

func NewEmailService(cfg config.EmailConfig) *EmailService {
	return &EmailService{
		cfg: cfg,
	}
}

type EmailData struct {
	To      string
	Subject string
	Body    string
	HTML    bool
}

func (e *EmailService) SendEmail(to, subject, body string, isHTML bool) error {
	if !e.cfg.Enabled {
		logger.Info("Email service disabled, skipping send", map[string]interface{}{
			"to":      to,
			"subject": subject,
		})
		return nil
	}

	// Email authentication
	auth := smtp.PlainAuth("", e.cfg.Username, e.cfg.Password, e.cfg.SMTPHost)

	// Determine content type
	contentType := "text/plain"
	if isHTML {
		contentType = "text/html"
	}

	// Construct email message
	msg := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: %s; charset=utf-8\r\n"+
			"\r\n"+
			"%s",
		e.cfg.From, to, subject, contentType, body,
	))

	// Send email
	addr := fmt.Sprintf("%s:%d", e.cfg.SMTPHost, e.cfg.SMTPPort)
	err := smtp.SendMail(addr, auth, e.cfg.From, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	logger.Info("Email sent successfully", map[string]interface{}{
		"to":      to,
		"subject": subject,
	})

	return nil
}

// Template methods
func (e *EmailService) SendInvoiceEmail(to, customerName, invoiceNumber string, amount float64, pdfURL string) error {
	subject := fmt.Sprintf("Invoice #%s - Your ISP Bill", invoiceNumber)

	body := fmt.Sprintf(`
		Dear %s,<br><br>
		
		Your invoice <strong>%s</strong> of <strong>KES %.2f</strong> is ready.<br>
		Due Date: 15th of this month<br><br>
		
		You can view and pay your invoice here:<br>
		<a href="%s">View Invoice</a><br><br>
		
		Thank you for choosing our service.<br><br>
		
		Regards,<br>
		Your ISP Team
	`, customerName, invoiceNumber, amount, pdfURL)

	return e.SendEmail(to, subject, body, true)
}

func (e *EmailService) SendPaymentConfirmationEmail(to, customerName, receipt string, amount float64) error {
	subject := "Payment Confirmation - ISP Services"

	body := fmt.Sprintf(`
		Dear %s,<br><br>
		
		We have received your payment of <strong>KES %.2f</strong>.<br>
		Receipt Number: <strong>%s</strong><br><br>
		
		Your service is active and running.<br><br>
		
		Thank you for your payment.<br><br>
		
		Regards,<br>
		Your ISP Team
	`, customerName, amount, receipt)

	return e.SendEmail(to, subject, body, true)
}

func (e *EmailService) SendSuspensionWarningEmail(to, customerName string, daysOverdue int) error {
	subject := "⚠️ Suspension Warning - Action Required"

	body := fmt.Sprintf(`
		Dear %s,<br><br>
		
		Your account is <strong>%d days</strong> overdue.<br>
		Your service will be suspended in <strong>24 hours</strong>.<br><br>
		
		Please make a payment immediately to avoid interruption.<br><br>
		
		Regards,<br>
		Your ISP Team
	`, customerName, daysOverdue)

	return e.SendEmail(to, subject, body, true)
}
