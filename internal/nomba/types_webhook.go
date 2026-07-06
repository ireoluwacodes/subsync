package nomba

// Webhook event types subscribed in the Nomba dashboard.
type WebhookEventType string

const (
	WebhookEventPaymentSuccess  WebhookEventType = "payment_success"
	WebhookEventPaymentFailed   WebhookEventType = "payment_failed"
	WebhookEventPaymentReversal WebhookEventType = "payment_reversal"
	WebhookEventPayoutSuccess   WebhookEventType = "payout_success"
	WebhookEventPayoutFailed    WebhookEventType = "payout_failed"
	WebhookEventPayoutRefund    WebhookEventType = "payout_refund"
)

// WebhookEvent is the inbound POST body from Nomba webhooks.
type WebhookEvent struct {
	EventType WebhookEventType `json:"event_type"`
	RequestID string           `json:"requestId"`
	Data      WebhookEventData `json:"data"`
}

type WebhookEventData struct {
	Merchant          WebhookMerchant          `json:"merchant"`
	Terminal          map[string]any           `json:"terminal,omitempty"`
	Transaction       WebhookTransaction       `json:"transaction"`
	Customer          WebhookCustomer          `json:"customer,omitempty"`
	Order             *WebhookOrder            `json:"order,omitempty"`
	TokenizedCardData *WebhookTokenizedCardData `json:"tokenizedCardData,omitempty"`
}

// WebhookOrder is present on online checkout payment_success events.
type WebhookOrder struct {
	OrderReference         string            `json:"orderReference,omitempty"`
	PaymentMethod          string            `json:"paymentMethod,omitempty"`
	OrderMetaData          map[string]string `json:"orderMetaData,omitempty"`
	CustomerEmail          string            `json:"customerEmail,omitempty"`
	IsTokenizedCardPayment string            `json:"isTokenizedCardPayment,omitempty"`
	CardLast4Digits        string            `json:"cardLast4Digits,omitempty"`
	CardType               string            `json:"cardType,omitempty"`
}

// WebhookTokenizedCardData may carry card token fields (often "N/A" for transfers).
type WebhookTokenizedCardData struct {
	TokenKey string `json:"tokenKey,omitempty"`
	CardType string `json:"cardType,omitempty"`
	CardPan  string `json:"cardPan,omitempty"`
}

type WebhookMerchant struct {
	WalletID      string  `json:"walletId,omitempty"`
	WalletBalance float64 `json:"walletBalance,omitempty"`
	UserID        string  `json:"userId,omitempty"`
}

type WebhookTransaction struct {
	TransactionID         string  `json:"transactionId"`
	Type                  string  `json:"type"`
	Time                  string  `json:"time"`
	ResponseCode          string  `json:"responseCode,omitempty"`
	ResponseCodeMessage   string  `json:"responseCodeMessage,omitempty"`
	TransactionAmount     float64 `json:"transactionAmount,omitempty"`
	Fee                   float64 `json:"fee,omitempty"`
	SessionID             string  `json:"sessionId,omitempty"`
	MerchantTxRef         string  `json:"merchantTxRef,omitempty"`
	Narration             string  `json:"narration,omitempty"`
	OriginatingFrom       string  `json:"originatingFrom,omitempty"`
	TokenKey              string  `json:"tokenKey,omitempty"`
	CardIssuer            string  `json:"cardIssuer,omitempty"`
	AliasAccountNumber    string  `json:"aliasAccountNumber,omitempty"`
	AliasAccountName      string  `json:"aliasAccountName,omitempty"`
	AliasAccountReference string  `json:"aliasAccountReference,omitempty"`
	AliasAccountType      string  `json:"aliasAccountType,omitempty"`
}

type WebhookCustomer struct {
	BankCode      string `json:"bankCode,omitempty"`
	SenderName    string `json:"senderName,omitempty"`
	RecipientName string `json:"recipientName,omitempty"`
	BankName      string `json:"bankName,omitempty"`
	AccountNumber string `json:"accountNumber,omitempty"`
	CardPan       string `json:"cardPan,omitempty"`
	CardIssuer    string `json:"cardIssuer,omitempty"`
	CardBank      string `json:"cardBank,omitempty"`
	BillerID      string `json:"billerId,omitempty"`
	ProductID     string `json:"productId,omitempty"`
}

// WebhookHeaders carries Nomba signature metadata from inbound webhook HTTP headers.
type WebhookHeaders struct {
	Signature          string
	SignatureValue     string
	SignatureAlgorithm string
	SignatureVersion   string
	Timestamp          string
}
