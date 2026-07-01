package nomba

// Currency codes supported on Nigerian checkout.
type Currency string

const (
	CurrencyNGN Currency = "NGN"
	CurrencyCDF Currency = "CDF"
	CurrencyUSD Currency = "USD"
)

// PaymentMethod values for order.allowedPaymentMethods.
type PaymentMethod string

const (
	PaymentMethodCard          PaymentMethod = "Card"
	PaymentMethodTransfer      PaymentMethod = "Transfer"
	PaymentMethodNombaQR       PaymentMethod = "Nomba QR"
	PaymentMethodUSSD          PaymentMethod = "USSD"
	PaymentMethodBuyNowPayLater PaymentMethod = "Buy Now Pay Later"
)

// SplitType for order.splitRequest.
type SplitType string

const (
	SplitTypePercentage SplitType = "PERCENTAGE"
	SplitTypeAmount     SplitType = "AMOUNT"
)

// Order is the checkout order object shared by create-order and tokenized-card-payment.
type Order struct {
	OrderReference        string            `json:"orderReference,omitempty"`
	CustomerID            string            `json:"customerId,omitempty"`
	CallbackURL           string            `json:"callbackUrl"`
	CustomerEmail         string            `json:"customerEmail"`
	Amount                float64           `json:"amount"`
	Currency              Currency          `json:"currency"`
	AccountID             string            `json:"accountId,omitempty"`
	AllowedPaymentMethods []PaymentMethod   `json:"allowedPaymentMethods,omitempty"`
	SplitRequest          *SplitRequest     `json:"splitRequest,omitempty"`
	OrderMetaData         map[string]string `json:"orderMetaData,omitempty"`
}

type SplitRequest struct {
	SplitType SplitType        `json:"splitType"`
	SplitList []SplitListEntry `json:"splitList"`
}

type SplitListEntry struct {
	AccountID string  `json:"accountId"`
	Value     float64 `json:"value"`
}

// CreateOrderRequest is the body for POST /v1/checkout/order.
type CreateOrderRequest struct {
	Order        Order `json:"order"`
	TokenizeCard bool  `json:"tokenizeCard,omitempty"`
}

// CreateOrderResult is the data payload from create-order.
type CreateOrderResult struct {
	CheckoutLink   string `json:"checkoutLink"`
	OrderReference string `json:"orderReference"`
}

// TokenizedCardPaymentRequest is the body for POST /v1/checkout/tokenized-card-payment.
type TokenizedCardPaymentRequest struct {
	Order    Order  `json:"order"`
	TokenKey string `json:"tokenKey"`
}

// TokenizedCardPaymentResult is the data payload from tokenized-card-payment.
type TokenizedCardPaymentResult struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

// OrderDetails is returned when fetching/verifying a checkout order.
type OrderDetails struct {
	OrderID        string   `json:"orderId,omitempty"`
	OrderReference string   `json:"orderReference,omitempty"`
	CustomerID     string   `json:"customerId,omitempty"`
	AccountID      string   `json:"accountId,omitempty"`
	CallbackURL    string   `json:"callbackUrl,omitempty"`
	CustomerEmail  string   `json:"customerEmail,omitempty"`
	Amount         float64  `json:"amount,omitempty"`
	Currency       Currency `json:"currency,omitempty"`
	BusinessName   string   `json:"businessName,omitempty"`
	BusinessEmail  string   `json:"businessEmail,omitempty"`
	BusinessLogo   string   `json:"businessLogo,omitempty"`
}

// CheckoutTransactionDetailsResult is the data payload from fetch checkout transaction.
type CheckoutTransactionDetailsResult struct {
	Status  bool          `json:"status"`
	Message string        `json:"message"`
	Order   *OrderDetails `json:"order,omitempty"`
}

// TokenizedCardData represents a stored card token (list/update/delete endpoints).
type TokenizedCardData struct {
	TokenKey   string `json:"tokenKey,omitempty"`
	CardLast4  string `json:"cardLast4,omitempty"`
	CardBrand  string `json:"cardBrand,omitempty"`
	CardExpiry string `json:"cardExpiry,omitempty"`
}
