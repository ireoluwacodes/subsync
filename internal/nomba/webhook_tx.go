package nomba

import "strings"

// IsTransferTransaction reports whether a Nomba payment_success transaction was paid via bank transfer.
func IsTransferTransaction(tx WebhookTransaction) bool {
	t := strings.ToLower(strings.TrimSpace(tx.Type))
	if t == "" {
		return tx.TokenKey == "" && (tx.AliasAccountReference != "" || tx.SessionID != "")
	}
	return strings.Contains(t, "transfer") || strings.Contains(t, "vact")
}

// IsCardTransaction reports whether the webhook indicates a card payment with a token.
func IsCardTransaction(tx WebhookTransaction) bool {
	if tx.TokenKey != "" {
		return true
	}
	t := strings.ToLower(strings.TrimSpace(tx.Type))
	return strings.Contains(t, "card") || strings.Contains(t, "purchase")
}

func ParsePaymentMethods(values []string) []PaymentMethod {
	if len(values) == 0 {
		return nil
	}
	out := make([]PaymentMethod, 0, len(values))
	for _, v := range values {
		switch strings.TrimSpace(v) {
		case string(PaymentMethodCard):
			out = append(out, PaymentMethodCard)
		case string(PaymentMethodTransfer):
			out = append(out, PaymentMethodTransfer)
		case string(PaymentMethodNombaQR):
			out = append(out, PaymentMethodNombaQR)
		case string(PaymentMethodUSSD):
			out = append(out, PaymentMethodUSSD)
		case string(PaymentMethodBuyNowPayLater):
			out = append(out, PaymentMethodBuyNowPayLater)
		}
	}
	return out
}
