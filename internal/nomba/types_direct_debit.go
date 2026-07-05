package nomba

// MandateFrequency for POST /v1/direct-debits.
type MandateFrequency string

const (
	MandateFrequencyVariable         MandateFrequency = "VARIABLE"
	MandateFrequencyWeekly           MandateFrequency = "WEEKLY"
	MandateFrequencyEveryTwoWeeks    MandateFrequency = "EVERY_TWO_WEEKS"
	MandateFrequencyMonthly          MandateFrequency = "MONTHLY"
	MandateFrequencyEveryTwoMonths   MandateFrequency = "EVERY_TWO_MONTHS"
	MandateFrequencyEveryThreeMonths MandateFrequency = "EVERY_THREE_MONTHS"
	MandateFrequencyEveryFourMonths  MandateFrequency = "EVERY_FOUR_MONTHS"
	MandateFrequencyEveryFiveMonths  MandateFrequency = "EVERY_FIVE_MONTHS"
	MandateFrequencyEverySixMonths   MandateFrequency = "EVERY_SIX_MONTHS"
	MandateFrequencyEverySevenMonths MandateFrequency = "EVERY_SEVEN_MONTHS"
	MandateFrequencyEveryEightMonths MandateFrequency = "EVERY_EIGHT_MONTHS"
	MandateFrequencyEveryNineMonths  MandateFrequency = "EVERY_NINE_MONTHS"
	MandateFrequencyEveryTenMonths   MandateFrequency = "EVERY_TEN_MONTHS"
	MandateFrequencyEveryElevenMonths MandateFrequency = "EVERY_ELEVEN_MONTHS"
	MandateFrequencyEveryTwelveMonths MandateFrequency = "EVERY_TWELVE_MONTHS"
)

// CreateMandateRequest is the body for POST /v1/direct-debits.
type CreateMandateRequest struct {
	CustomerAccountNumber string           `json:"customerAccountNumber"`
	BankCode              string           `json:"bankCode"`
	CustomerName          string           `json:"customerName"`
	CustomerAddress       string           `json:"customerAddress,omitempty"`
	CustomerAccountName   string           `json:"customerAccountName"`
	Amount                float64          `json:"amount"`
	Frequency             MandateFrequency `json:"frequency"`
	Narration             string           `json:"narration,omitempty"`
	CustomerPhoneNumber   string           `json:"customerPhoneNumber,omitempty"`
	MerchantReference     string           `json:"merchantReference"`
	StartDate             string           `json:"startDate"`
	EndDate               string           `json:"endDate"`
	CustomerEmail         string           `json:"customerEmail"`
	StartImmediately      bool             `json:"startImmediately,omitempty"`
}

// CreateMandateResult is the data payload from mandate creation.
type CreateMandateResult struct {
	MandateID         string `json:"mandateId"`
	MerchantReference string `json:"merchantReference"`
	PhoneNumber       string `json:"phoneNumber"`
	Description       string `json:"description"`
}

// MandateStatusResult is the data payload from GET /v1/direct-debits/status.
type MandateStatusResult struct {
	CustomerAccountName   string `json:"customerAccountName"`
	MandateID             string `json:"mandateId"`
	CustomerAccountNumber string `json:"customerAccountNumber"`
	MandateStatus         string `json:"mandateStatus"`
	RejectionComment      string `json:"rejectionComment,omitempty"`
	MandateAdviceStatus   string `json:"mandateAdviceStatus,omitempty"`
}

// MandateReadyForDebit reports whether a mandate can be debited per SubSync dunning rules.
func (m MandateStatusResult) MandateReadyForDebit() bool {
	return m.MandateStatus == MandateStatusActive &&
		m.MandateAdviceStatus == MandateAdviceSent
}

// DebitMandateRequest is the body for POST /v1/direct-debits/debit-mandate.
type DebitMandateRequest struct {
	MandateID         string `json:"mandateId"`
	Amount            string `json:"amount"`
	MerchantReference string `json:"merchantReference,omitempty"`
}

// DebitMandateResult is the data payload from debit-mandate.
type DebitMandateResult struct {
	MandateID string `json:"mandateId"`
	Status    string `json:"status"`
	Amount    string `json:"amount"`
	Message   string `json:"message"`
}
