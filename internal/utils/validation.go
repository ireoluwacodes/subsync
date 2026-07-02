package utils

import (
	"fmt"
	"strings"

	"github.com/ireoluwacodes/subsync/internal/domain"
)

func ValidateNombaInput(clientID, clientSecret, accountID, nombaEnv string) error {
	if clientID == "" || clientSecret == "" || accountID == "" {
		return fmt.Errorf("%w: nomba_client_id, nomba_client_secret, and nomba_account_id are required", domain.ErrValidation)
	}
	env := strings.ToLower(nombaEnv)
	if env != domain.NombaEnvSandbox && env != domain.NombaEnvProduction {
		return fmt.Errorf("%w: nomba_env must be sandbox or production", domain.ErrValidation)
	}
	return nil
}

func ValidatePlanInput(interval domain.PlanInterval, intervalDays *int, amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("%w: amount must be greater than zero", domain.ErrValidation)
	}
	switch interval {
	case domain.PlanIntervalMonthly, domain.PlanIntervalAnnual:
		return nil
	case domain.PlanIntervalCustom:
		if intervalDays == nil || *intervalDays <= 0 {
			return fmt.Errorf("%w: interval_days is required for custom interval", domain.ErrValidation)
		}
		return nil
	default:
		return fmt.Errorf("%w: invalid interval", domain.ErrValidation)
	}
}

func ValidatePaymentMethod(pmType domain.PaymentMethodType, tokenKey, mandateID string) error {
	switch pmType {
	case domain.PaymentMethodTokenizedCard:
		if tokenKey == "" {
			return fmt.Errorf("%w: token_key is required for tokenized_card", domain.ErrValidation)
		}
	case domain.PaymentMethodDirectDebit:
		if mandateID == "" {
			return fmt.Errorf("%w: mandate_id is required for direct_debit", domain.ErrValidation)
		}
	default:
		return fmt.Errorf("%w: invalid payment method type", domain.ErrValidation)
	}
	return nil
}
