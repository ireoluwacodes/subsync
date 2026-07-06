package nomba

// TransferStatus values returned by bank transfer endpoints.
type TransferStatus string

const (
	TransferStatusSuccess        TransferStatus = "SUCCESS"
	TransferStatusPendingBilling TransferStatus = "PENDING_BILLING"
)

// TransferType classifies a transfer transaction.
type TransferType string

const (
	TransferTypeWithdrawal      TransferType = "withdrawal"
	TransferTypePurchase        TransferType = "purchase"
	TransferTypeTransfer        TransferType = "transfer"
	TransferTypeP2P             TransferType = "p2p"
	TransferTypeOnlineCheckout  TransferType = "online_checkout"
	TransferTypeQRTCredit       TransferType = "qrt_credit"
	TransferTypeQRTDebit        TransferType = "qrt_debit"
)

// BankAccountTransferRequest is the body for POST /v2/transfers/bank/{subAccountId}.
type BankAccountTransferRequest struct {
	Amount        float64 `json:"amount"`
	AccountNumber string  `json:"accountNumber"`
	AccountName   string  `json:"accountName"`
	BankCode      string  `json:"bankCode"`
	MerchantTxRef string  `json:"merchantTxRef"`
	SenderName    string  `json:"senderName,omitempty"`
	Narration     string  `json:"narration,omitempty"`
}

// BankAccountTransferResult is the data payload from sub-account bank transfer.
type BankAccountTransferResult struct {
	Amount           string                     `json:"amount"`
	Source           string                     `json:"source,omitempty"`
	SourceUserID     string                     `json:"sourceUserId,omitempty"`
	CustomerBillerID string                     `json:"customerBillerId,omitempty"`
	ProductID        string                     `json:"productId,omitempty"`
	Meta             BankAccountTransferMeta    `json:"meta"`
	Fee              float64                    `json:"fee"`
	TimeCreated      string                     `json:"timeCreated"`
	ID               string                     `json:"id"`
	Type             TransferType               `json:"type"`
	Status           TransferStatus             `json:"status"`
}

type BankAccountTransferMeta struct {
	APIRRN              string `json:"api_rrn,omitempty"`
	Narration           string `json:"narration,omitempty"`
	RecipientName       string `json:"recipientName,omitempty"`
	SenderName          string `json:"sender_name,omitempty"`
	MerchantTxRef       string `json:"merchantTxRef,omitempty"`
	APIClientID         string `json:"api_client_id,omitempty"`
	Currency            string `json:"currency,omitempty"`
	HooksEligible       string `json:"hooksEligible,omitempty"`
	BankingEntityID     string `json:"banking_entity_id,omitempty"`
	BankingEntityUserID string `json:"banking_entity_user_id,omitempty"`
	BankingEntityType   string `json:"banking_entity_type,omitempty"`
	SelfTransaction     string `json:"self_transaction,omitempty"`
	TransactionCategory string `json:"transactionCategory,omitempty"`
	AccountNumber       string `json:"accountNumber,omitempty"`
	BankName            string `json:"bankName,omitempty"`
	BankCode            string `json:"bankCode,omitempty"`
}

// AccountBalanceResult is the data payload from fetch sub-account balance.
type AccountBalanceResult struct {
	Amount      string   `json:"amount"`
	Currency    Currency `json:"currency"`
	TimeCreated string   `json:"timeCreated"`
}

// BanksListResults is the data payload from GET /v1/transfers/banks.
type BanksListResults struct {
	Results []Bank `json:"results"`
}

// Bank is a Nigerian bank from GET /v1/transfers/banks.
type Bank struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	NIPCode string `json:"nipCode,omitempty"`
	Logo    string `json:"logo,omitempty"`
}

// BankAccountLookupRequest is the body for bank account name lookup.
type BankAccountLookupRequest struct {
	AccountNumber string `json:"accountNumber"`
	BankCode      string `json:"bankCode"`
}

// BankAccountLookupResult is the data payload from account lookup.
type BankAccountLookupResult struct {
	AccountName   string `json:"accountName"`
	AccountNumber string `json:"accountNumber"`
	BankCode      string `json:"bankCode"`
}
