package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/clock"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/nomba"
)

const (
	mandateEndYears        = 5
	mandateStartLeadMinutes = 2
)

var nigeriaLoc = func() *time.Location {
	loc, err := time.LoadLocation("Africa/Lagos")
	if err != nil {
		return time.FixedZone("WAT", 1*60*60)
	}
	return loc
}()

// mandateScheduleTimes formats start/end for Nomba direct debit in Nigeria local time (WAT).
// Nomba validates startDate against local time; UTC timestamps can appear in the past.
func mandateScheduleTimes(now time.Time) (start, end string) {
	const layout = "2006-01-02T15:04"
	startAt := now.In(nigeriaLoc).Add(mandateStartLeadMinutes * time.Minute)
	endAt := startAt.AddDate(mandateEndYears, 0, 0)
	return startAt.Format(layout), endAt.Format(layout)
}

type MandateService struct {
	clock    clock.Clock
	repos    *db.Repos
	nomba    *nomba.Client
	webhooks *WebhookService
}

func NewMandateService(clk clock.Clock, repos *db.Repos, nombaClient *nomba.Client, webhooks *WebhookService) *MandateService {
	if clk == nil {
		clk = clock.RealClock{}
	}
	return &MandateService{clock: clk, repos: repos, nomba: nombaClient, webhooks: webhooks}
}

type DirectDebitSetupInput struct {
	CustomerAccountNumber string
	BankCode              string
	CustomerName          string
	CustomerAccountName   string
	CustomerEmail         string
	CustomerPhone         string
	CustomerAddress       string
}

type DirectDebitSetupResult struct {
	MandateID      string `json:"mandate_id"`
	Description    string `json:"description"`
	PhoneNumber    string `json:"phone_number"`
	MandateStatus  string `json:"mandate_status"`
	PaymentMethodID uuid.UUID `json:"payment_method_id"`
}

type DirectDebitStatusResult struct {
	MandateID       string `json:"mandate_id"`
	MandateStatus   string `json:"mandate_status"`
	NombaStatus     string `json:"nomba_status,omitempty"`
	AdviceStatus    string `json:"advice_status,omitempty"`
	Ready           bool   `json:"ready"`
	Instructions    string `json:"instructions,omitempty"`
}

func (s *MandateService) CreateForSubscription(
	ctx context.Context,
	tenant *domain.Tenant,
	sub *domain.Subscription,
	plan *domain.Plan,
	customer *domain.Customer,
	in DirectDebitSetupInput,
) (*DirectDebitSetupResult, error) {
	if err := s.repos.Tenants.LoadNombaSecret(ctx, tenant); err != nil {
		return nil, err
	}
	now := s.clock.Now().UTC()
	startDate, endDate := mandateScheduleTimes(now)
	result, err := s.nomba.CreateMandate(ctx, tenant, nomba.CreateMandateRequest{
		CustomerAccountNumber: in.CustomerAccountNumber,
		BankCode:              in.BankCode,
		CustomerName:          in.CustomerName,
		CustomerAccountName:   in.CustomerAccountName,
		CustomerAddress:       in.CustomerAddress,
		Amount:                float64(plan.Amount) / 100.0,
		Frequency:             nomba.MandateFrequencyVariable,
		Narration:             fmt.Sprintf("SubSync subscription %s", plan.Name),
		CustomerPhoneNumber:   in.CustomerPhone,
		MerchantReference:     numericMerchantRef(sub.ID, now),
		StartDate:             startDate,
		EndDate:               endDate,
		CustomerEmail:         in.CustomerEmail,
		StartImmediately:      true,
	})
	if err != nil {
		return nil, err
	}

	pm := &domain.PaymentMethod{
		TenantID:      tenant.ID,
		CustomerID:    customer.ID,
		Type:          domain.PaymentMethodDirectDebit,
		MandateID:     result.MandateID,
		MandateStatus: domain.MandateStatusPending,
		IsDefault:     false,
	}
	if err := s.repos.PaymentMethods.Create(ctx, pm); err != nil {
		return nil, err
	}

	sub.FallbackPaymentMethodID = &pm.ID
	setSubscriptionMeta(sub, domain.SubscriptionMetaMandateInstructions, result.Description)
	if err := s.repos.Subscriptions.Update(ctx, sub); err != nil {
		return nil, err
	}

	return &DirectDebitSetupResult{
		MandateID:       result.MandateID,
		Description:     result.Description,
		PhoneNumber:     result.PhoneNumber,
		MandateStatus:   string(domain.MandateStatusPending),
		PaymentMethodID: pm.ID,
	}, nil
}

func (s *MandateService) RefreshStatus(ctx context.Context, tenant *domain.Tenant, pm *domain.PaymentMethod) (*DirectDebitStatusResult, error) {
	if pm == nil || pm.MandateID == "" {
		return nil, fmt.Errorf("%w: mandate not found", domain.ErrNotFound)
	}
	if err := s.repos.Tenants.LoadNombaSecret(ctx, tenant); err != nil {
		return nil, err
	}
	status, err := s.nomba.GetMandateStatus(ctx, tenant, pm.MandateID)
	if err != nil {
		return nil, err
	}
	out := &DirectDebitStatusResult{
		MandateID:     pm.MandateID,
		MandateStatus: string(pm.MandateStatus),
		NombaStatus:   status.MandateStatus,
		AdviceStatus:  status.MandateAdviceStatus,
		Ready:         status.MandateReadyForDebit(),
	}
	if status.MandateReadyForDebit() && pm.MandateStatus != domain.MandateStatusReady {
		if err := s.activateMandatePM(ctx, tenant.ID, pm); err != nil {
			return nil, err
		}
		out.MandateStatus = string(domain.MandateStatusReady)
		out.Ready = true
	}
	return out, nil
}

func (s *MandateService) ProcessPendingMandates(ctx context.Context, limit int) (int, error) {
	pms, err := s.repos.PaymentMethods.ListPendingMandates(ctx, limit)
	if err != nil {
		return 0, err
	}
	processed := 0
	for _, pm := range pms {
		tenant, err := s.repos.Tenants.GetByID(ctx, pm.TenantID)
		if err != nil {
			continue
		}
		if _, err := s.RefreshStatus(ctx, tenant, pm); err != nil {
			continue
		}
		processed++
	}
	return processed, nil
}

func (s *MandateService) activateMandatePM(ctx context.Context, tenantID uuid.UUID, pm *domain.PaymentMethod) error {
	pm.MandateStatus = domain.MandateStatusReady
	if err := s.repos.PaymentMethods.Update(ctx, pm); err != nil {
		return err
	}

	subs, _, err := s.repos.Subscriptions.List(ctx, tenantID, domain.SubscriptionListFilter{
		CustomerID: &pm.CustomerID,
		Limit:      100,
	})
	if err != nil {
		return err
	}
	for _, sub := range subs {
		if sub.FallbackPaymentMethodID == nil || *sub.FallbackPaymentMethodID != pm.ID {
			continue
		}
		setSubscriptionMeta(sub, domain.SubscriptionMetaAwaitingPaymentMethod, nil)
		setSubscriptionMeta(sub, domain.SubscriptionMetaMandateInstructions, nil)
		clearPMReminderMetadata(sub)
		if err := s.repos.Subscriptions.Update(ctx, sub); err != nil {
			return err
		}
		if s.webhooks != nil {
			_ = s.webhooks.Emit(ctx, tenantID, domain.WebhookEventPaymentMethodAttached, map[string]any{
				"id":          pm.ID.String(),
				"customer_id": pm.CustomerID.String(),
				"type":        string(pm.Type),
			})
		}
	}
	return nil
}

func numericMerchantRef(subID uuid.UUID, now time.Time) string {
	var n uint64
	for i := 0; i < 8; i++ {
		n = n*256 + uint64(subID[i])
	}
	return fmt.Sprintf("%010d%010d", n%1_000_000_0000, uint64(now.Unix())%1_000_000_0000)
}
