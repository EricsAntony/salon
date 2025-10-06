package gateway

import (
	"fmt"

	"payment-service/internal/config"
	"payment-service/internal/model"
)

// gatewayManager implements GatewayManager interface
type gatewayManager struct {
	gateways map[string]PaymentGateway
	config   *config.Config
}

// NewGatewayManager creates a new gateway manager
func NewGatewayManager(cfg *config.Config) GatewayManager {
	manager := &gatewayManager{
		gateways: make(map[string]PaymentGateway),
		config:   cfg,
	}

	// Initialize available gateways based on configuration
	if cfg.StripeSecretKey != "" {
		stripeGateway := NewStripeGateway(cfg.StripeSecretKey, cfg.StripeWebhookSecret)
		manager.gateways[model.GatewayStripe] = stripeGateway
	}

	if cfg.RazorpayKeyID != "" && cfg.RazorpayKeySecret != "" {
		razorpayGateway := NewRazorpayGateway(cfg.RazorpayKeyID, cfg.RazorpayKeySecret, cfg.RazorpayWebhookSecret)
		manager.gateways[model.GatewayRazorpay] = razorpayGateway
	}

	return manager
}

// GetGateway returns a specific payment gateway by name
func (m *gatewayManager) GetGateway(name string) (PaymentGateway, error) {
	gateway, exists := m.gateways[name]
	if !exists {
		return nil, fmt.Errorf("payment gateway '%s' not found or not configured", name)
	}
	return gateway, nil
}

// GetAvailableGateways returns list of available gateway names
func (m *gatewayManager) GetAvailableGateways() []string {
	var gateways []string
	for name := range m.gateways {
		gateways = append(gateways, name)
	}
	return gateways
}

// SelectBestGateway selects the best gateway based on amount and currency
func (m *gatewayManager) SelectBestGateway(amount float64, currency string) (PaymentGateway, error) {
	// Gateway selection logic based on business rules
	
	// For INR currency, prefer Razorpay
	if currency == "INR" {
		if gateway, exists := m.gateways[model.GatewayRazorpay]; exists {
			return gateway, nil
		}
	}

	// For international currencies, prefer Stripe
	if currency != "INR" {
		if gateway, exists := m.gateways[model.GatewayStripe]; exists {
			return gateway, nil
		}
	}

	// Fallback to any available gateway
	for _, gateway := range m.gateways {
		return gateway, nil
	}

	return nil, fmt.Errorf("no payment gateways available")
}
