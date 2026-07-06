package nomba

const (
	PathAuthTokenIssue    = "/v1/auth/token/issue"
	PathAuthTokenRefresh  = "/v1/auth/token/refresh"
	PathAuthTokenRevoke   = "/v1/auth/token/revoke"
	PathCheckoutOrder     = "/v1/checkout/order"
	PathTokenizedPayment  = "/v1/checkout/tokenized-card-payment"
	PathCheckoutVerify    = "/v1/checkout/transactions"
	PathDirectDebitCreate = "/v1/direct-debits"
	PathDirectDebitStatus = "/v1/direct-debits/status"
	PathDirectDebitDebit   = "/v1/direct-debits/debit-mandate"
	PathTransfersBanks     = "/v1/transfers/banks"
	PathTransfersBankLookup = "/v1/transfers/bank/lookup"
	PathSubAccountTransfer = "/v2/transfers/bank/%s" // fmt with subAccountId
)

// HeaderAccountID is sent on every authenticated Nomba request (parent account ID).
const HeaderAccountID = "accountId"

// Webhook header names (case-insensitive per Nomba docs).
const (
	HeaderNombaSignature          = "nomba-signature"
	HeaderNombaSignatureValue     = "nomba-sig-value"
	HeaderNombaSignatureAlgorithm = "nomba-signature-algorithm"
	HeaderNombaTimestamp          = "nomba-timestamp"
)

// Mandate status values from GET /v1/direct-debits/status.
const (
	MandateStatusActive    = "Active"
	MandateAdviceSent      = "ADVICE_SENT"
	MandateAdviceNotSent   = "ADVICE_NOT_SENT"
	mandateAdviceSentLegacy = "Advice sent" // older Nomba docs / sandbox responses
)

// Response success code.
const ResponseCodeSuccess = "00"
