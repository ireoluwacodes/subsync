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

type Renderer struct {
	layout  *template.Template
	home    *template.Template
	ddForm  *template.Template
	ddPending *template.Template
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
	return &Renderer{layout: layout, home: home, ddForm: ddForm, ddPending: ddPending}, nil
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
	Token         string
	TenantName    string
	PlanName      string
	CustomerEmail string
	CustomerName  string
	FlashError    string
}

type DirectDebitPendingData struct {
	Title         string
	Token         string
	TenantName    string
	Instructions  string
	MandateStatus string
	Ready         bool
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
