package portalpage

import (
	_ "embed"
	"html/template"
	"io"
)

//go:embed templates/layout.html
var layoutHTML string

//go:embed templates/home.html
var homeHTML string

//go:embed templates/direct_debit.html
var directDebitHTML string

//go:embed templates/direct_debit_pending.html
var directDebitPendingHTML string

//go:embed templates/billing_success.html
var billingSuccessHTML string

type Renderer struct {
	layout        *template.Template
	home          *template.Template
	ddForm        *template.Template
	ddPending     *template.Template
	billingSuccess *template.Template
}

func NewRenderer() (*Renderer, error) {
	layout, err := template.New("layout").Parse(layoutHTML)
	if err != nil {
		return nil, err
	}
	home, err := layout.Clone()
	if err != nil {
		return nil, err
	}
	if _, err = home.Parse(homeHTML); err != nil {
		return nil, err
	}
	ddForm, err := layout.Clone()
	if err != nil {
		return nil, err
	}
	if _, err = ddForm.Parse(directDebitHTML); err != nil {
		return nil, err
	}
	ddPending, err := layout.Clone()
	if err != nil {
		return nil, err
	}
	if _, err = ddPending.Parse(directDebitPendingHTML); err != nil {
		return nil, err
	}
	billingSuccess, err := layout.Clone()
	if err != nil {
		return nil, err
	}
	if _, err = billingSuccess.Parse(billingSuccessHTML); err != nil {
		return nil, err
	}
	return &Renderer{layout: layout, home: home, ddForm: ddForm, ddPending: ddPending, billingSuccess: billingSuccess}, nil
}

type HomeData struct {
	Title                  string
	Token                  string
	TenantName             string
	PlanName               string
	CustomerEmail          string
	State                  string
	CancelAtPeriodEnd      bool
	AwaitingPaymentMethod  bool
	HasCard                bool
	HasMandate             bool
	MandateStatus          string
	CanSetupDirectDebit    bool
	PaymentMethodLast4     string
	PaymentMethodBrand     string
	FlashMessage           string
	FlashError             string
}

type DirectDebitFormData struct {
	Title         string
	TenantName    string
	FlashMessage  string
	FlashError    string
	Token         string
	PlanName      string
	CustomerEmail string
	CustomerName  string
}

type DirectDebitPendingData struct {
	Title         string
	TenantName    string
	FlashMessage  string
	FlashError    string
	Token         string
	Instructions  string
	MandateStatus string
	Ready         bool
}

type BillingSuccessData struct {
	Title             string
	TenantName        string
	FlashMessage      string
	FlashError        string
	PlanName          string
	StatusLabel      string
	StatusMessage    string
	Outcome          string
	OutcomeBadge     string
	OrderReference   string
	OrderID          string
	SubscriptionState string
	AmountDisplay    string
}

func (r *Renderer) RenderHome(w io.Writer, data HomeData) error {
	return r.home.ExecuteTemplate(w, "layout", data)
}

func (r *Renderer) RenderDirectDebitForm(w io.Writer, data DirectDebitFormData) error {
	return r.ddForm.ExecuteTemplate(w, "layout", data)
}

func (r *Renderer) RenderDirectDebitPending(w io.Writer, data DirectDebitPendingData) error {
	return r.ddPending.ExecuteTemplate(w, "layout", data)
}

func (r *Renderer) RenderBillingSuccess(w io.Writer, data BillingSuccessData) error {
	return r.billingSuccess.ExecuteTemplate(w, "layout", data)
}
