package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/your-org/isp-billing/internal/config"
	"github.com/your-org/isp-billing/internal/domain"
	"github.com/your-org/isp-billing/internal/service/payment"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env file not found, using existing environment variables")
	}

	// Manual .env loading as fallback
	loadEnvFile()

	cfg := config.LoadFromEnv()

	fmt.Println("========================================")
	fmt.Println("M-PESA Sandbox Test")
	fmt.Println("========================================")
	fmt.Printf("Consumer Key: %s\n", maskString(cfg.MPESA.ConsumerKey))
	fmt.Printf("Consumer Secret: %s\n", maskString(cfg.MPESA.ConsumerSecret))
	fmt.Printf("Environment: %s\n", cfg.MPESA.Environment)
	fmt.Printf("Shortcode: %s\n", cfg.MPESA.ShortCode)
	fmt.Printf("Passkey: %s\n", func() string {
		if cfg.MPESA.Passkey != "" && cfg.MPESA.Passkey != "your_passkey_here" {
			return "✅ Configured"
		}
		return "❌ NOT CONFIGURED - REQUIRED!"
	}())
	fmt.Println("========================================\n")

	// Check if Passkey is configured
	if cfg.MPESA.Passkey == "" || cfg.MPESA.Passkey == "your_passkey_here" {
		fmt.Println("❌ ERROR: M-PESA Passkey is not configured!")
		fmt.Println("\nPlease update .env with your actual Passkey")
		fmt.Println("Get it from: https://developer.safaricom.co.ke/")
		fmt.Println("  My Apps → Extreme Solutions Sandbox → M-PESA EXPRESS Sandbox")
		fmt.Println("\nThen update .env:")
		fmt.Println("  MPESA_PASSKEY=your_actual_passkey_here")
		return
	}

	if cfg.MPESA.ConsumerKey == "" || cfg.MPESA.ConsumerSecret == "" {
		fmt.Println("❌ ERROR: M-PESA Consumer Key/Secret not configured!")
		return
	}

	// Initialize M-PESA service
	mpesaService := payment.NewMPESAService(cfg.MPESA)

	fmt.Println("📱 Testing STK Push...")
	fmt.Println("\n⚠️  Using test phone: 254712345678")
	fmt.Println("   To change: export TEST_PHONE=2547XXXXXXXX")
	fmt.Println("   Or update the phone number in the script")
	fmt.Println("")

	testPhone := "254712345678"
	if envPhone := os.Getenv("TEST_PHONE"); envPhone != "" {
		testPhone = envPhone
	}

	req := &domain.PaymentInitiateRequest{
		InvoiceID:   fmt.Sprintf("TEST-%d", time.Now().Unix()),
		PhoneNumber: testPhone,
		Amount:      1.00,
	}

	fmt.Printf("📱 Phone: %s\n", req.PhoneNumber)
	fmt.Printf("💰 Amount: KES %.2f\n", req.Amount)
	fmt.Printf("📄 Invoice: %s\n", req.InvoiceID)
	fmt.Println("")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Initiate STK Push
	resp, err := mpesaService.InitiateSTKPush(ctx, req)
	if err != nil {
		log.Printf("❌ STK Push failed: %v", err)
		log.Println("\n📋 Troubleshooting tips:")
		log.Println("1. Make sure you have a valid Passkey")
		log.Println("2. Phone number format: 2547XXXXXXXX (no +)")
		log.Println("3. Check if the phone is registered with M-PESA")
		log.Println("4. Verify Consumer Key and Consumer Secret")
		log.Println("5. Check internet connectivity")
		return
	}

	fmt.Println("✅ STK Push initiated successfully!")
	fmt.Printf("📋 Checkout Request ID: %s\n", resp.CheckoutRequestID)
	fmt.Printf("📋 Merchant Request ID: %s\n", resp.MerchantRequestID)
	fmt.Printf("💬 Customer Message: %s\n", resp.CustomerMessage)

	fmt.Println("\n⏳ Waiting for customer to complete payment...")
	fmt.Println("📱 Check your phone for the M-PESA prompt")

	// Wait and check status
	for i := 0; i < 6; i++ {
		time.Sleep(10 * time.Second)
		status, err := mpesaService.QueryStatus(ctx, resp.CheckoutRequestID)
		if err != nil {
			log.Printf("⚠️  Query failed: %v", err)
			continue
		}

		fmt.Printf("📊 Status: %s\n", status)

		if status == "completed" {
			fmt.Println("\n✅ Payment completed successfully!")
			break
		} else if status == "failed" {
			fmt.Println("\n❌ Payment failed")
			break
		}
	}

	fmt.Println("\n========================================")
	fmt.Println("Test completed!")
	fmt.Println("========================================")
}

func maskString(s string) string {
	if len(s) <= 10 {
		return s
	}
	return s[:10] + "..."
}

func loadEnvFile() {
	data, err := os.ReadFile(".env")
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			value = strings.Trim(value, "\"'")
			os.Setenv(key, value)
		}
	}
}
