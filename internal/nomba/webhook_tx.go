package nomba

import "strings"

// IsPlaceholderToken reports Nomba sentinel values that mean no card token.
func IsPlaceholderToken(value string) bool {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "", "N/A", "NA", "NULL", "NONE":
		return true
	default:
		return false
	}
}

// CheckoutOrderReference returns the SubSync/Nomba order reference used to match invoices.
// Online checkout webhooks put the checkout order ref on data.order, not transaction.merchantTxRef.
func CheckoutOrderReference(tx WebhookTransaction, order *WebhookOrder) string {
	if order != nil {
		if ref := strings.TrimSpace(order.OrderReference); ref != "" {
			return ref
		}
	}
	if ref := strings.TrimSpace(tx.MerchantTxRef); ref != "" {
		return ref
	}
	return strings.TrimSpace(tx.AliasAccountReference)
}

// EffectiveTokenKey returns a usable card token from webhook fields, ignoring Nomba placeholders.
func EffectiveTokenKey(tx WebhookTransaction, tokenized *WebhookTokenizedCardData) string {
	if !IsPlaceholderToken(tx.TokenKey) {
		return strings.TrimSpace(tx.TokenKey)
	}
	if tokenized != nil && !IsPlaceholderToken(tokenized.TokenKey) {
		return strings.TrimSpace(tokenized.TokenKey)
	}
	return ""
}

// IsTransferPayment reports bank-transfer checkout from transaction type and/or order.paymentMethod.
func IsTransferPayment(tx WebhookTransaction, order *WebhookOrder) bool {
	if order != nil {
		switch strings.ToLower(strings.TrimSpace(order.PaymentMethod)) {
		case "bank_transfer", "transfer":
			return true
		}
	}
	return IsTransferTransaction(tx)
}

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
	if !IsPlaceholderToken(tx.TokenKey) {
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
