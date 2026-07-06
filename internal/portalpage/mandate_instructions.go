package portalpage

import (
	"fmt"
	"regexp"
	"strings"
)

type MandatePaymentOption struct {
	Label         string
	AccountNumber string
	Bank          string
	AccountName   string
}

type MandateInstructionsView struct {
	AmountDisplay  string
	PaymentOptions []MandatePaymentOption
	RawFallback    string
}

func (v MandateInstructionsView) Parsed() bool {
	return len(v.PaymentOptions) > 0
}

var (
	mandateAmountRE     = regexp.MustCompile(`(?i)(?:₦|N)\s*([0-9]+(?:\.[0-9]{2})?)`)
	mandateAccountNumRE = regexp.MustCompile(`(?i)account number:\s*([0-9]+)`)
	mandateBankRE       = regexp.MustCompile(`(?i)bank:\s*(.+?)\s+account name:`)
	mandateAcctNameRE   = regexp.MustCompile(`(?i)account name:\s*(.+)$`)
	mandateOrSplitRE    = regexp.MustCompile(`(?i)\s+OR\s+`)
	mandateDotSpaceRE   = regexp.MustCompile(`\.([A-Z])`)
)

// ParseMandateInstructions turns Nomba mandate description text into display-friendly fields.
func ParseMandateInstructions(raw string) MandateInstructionsView {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return MandateInstructionsView{AmountDisplay: "₦50.00"}
	}

	fixed := mandateDotSpaceRE.ReplaceAllString(raw, ". $1")
	amount := extractMandateAmount(fixed)

	idx := strings.Index(strings.ToLower(fixed), "account number:")
	if idx < 0 {
		return MandateInstructionsView{AmountDisplay: amount, RawFallback: raw}
	}

	optionsPart := strings.TrimSpace(fixed[idx:])
	segments := mandateOrSplitRE.Split(optionsPart, -1)
	options := make([]MandatePaymentOption, 0, len(segments))
	for _, segment := range segments {
		if opt, ok := parseMandatePaymentOption(segment); ok {
			options = append(options, opt)
		}
	}
	if len(options) == 0 {
		return MandateInstructionsView{AmountDisplay: amount, RawFallback: raw}
	}
	for i := range options {
		if len(options) > 1 {
			options[i].Label = fmt.Sprintf("Option %d", i+1)
		}
	}
	return MandateInstructionsView{
		AmountDisplay:  amount,
		PaymentOptions: options,
	}
}

func extractMandateAmount(text string) string {
	if m := mandateAmountRE.FindStringSubmatch(text); len(m) >= 2 {
		if strings.Contains(m[0], ".") {
			return "₦" + m[1]
		}
		return "₦" + m[1] + ".00"
	}
	return "₦50.00"
}

func parseMandatePaymentOption(segment string) (MandatePaymentOption, bool) {
	num := mandateAccountNumRE.FindStringSubmatch(segment)
	bank := mandateBankRE.FindStringSubmatch(segment)
	name := mandateAcctNameRE.FindStringSubmatch(segment)
	if len(num) < 2 || len(bank) < 2 || len(name) < 2 {
		return MandatePaymentOption{}, false
	}
	return MandatePaymentOption{
		AccountNumber: num[1],
		Bank:          strings.TrimSpace(bank[1]),
		AccountName:   strings.TrimSpace(name[1]),
	}, true
}
