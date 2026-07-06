package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/stretchr/testify/require"
)

type memoryInvoiceRepo struct {
	invoices []*domain.Invoice
}

func (m *memoryInvoiceRepo) Create(ctx context.Context, invoice *domain.Invoice) error {
	invoice.ID = uuid.New()
	m.invoices = append(m.invoices, invoice)
	return nil
}
func (m *memoryInvoiceRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Invoice, error) {
	for _, inv := range m.invoices {
		if inv.ID == id && inv.TenantID == tenantID {
			return inv, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (m *memoryInvoiceRepo) GetByNombaOrderRef(ctx context.Context, tenantID uuid.UUID, orderRef string) (*domain.Invoice, error) {
	return nil, domain.ErrNotFound
}
func (m *memoryInvoiceRepo) GetByNombaTransactionID(ctx context.Context, tenantID uuid.UUID, transactionID string) (*domain.Invoice, error) {
	return nil, domain.ErrNotFound
}
func (m *memoryInvoiceRepo) GetOpenBySubscription(ctx context.Context, tenantID, subscriptionID uuid.UUID) (*domain.Invoice, error) {
	return nil, domain.ErrNotFound
}
func (m *memoryInvoiceRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.InvoiceListFilter) ([]*domain.Invoice, int64, error) {
	return nil, 0, nil
}
func (m *memoryInvoiceRepo) Update(ctx context.Context, invoice *domain.Invoice) error { return nil }
func (m *memoryInvoiceRepo) CreateLineItem(ctx context.Context, item *domain.InvoiceLineItem) error {
	return nil
}
func (m *memoryInvoiceRepo) ListLineItems(ctx context.Context, tenantID, invoiceID uuid.UUID) ([]*domain.InvoiceLineItem, error) {
	return nil, nil
}
func (m *memoryInvoiceRepo) SumPaidByCustomer(ctx context.Context, tenantID, customerID uuid.UUID) (int64, error) {
	return 0, nil
}

func TestInvoiceService_MockChargeSuccess(t *testing.T) {
	repo := &memoryInvoiceRepo{}
	cfg := &config.Config{BillingMockResult: "success"}
	svc := NewInvoiceService(repo, nil, cfg, nil, nil, nil)
	inv := &domain.Invoice{
		TenantID:  uuid.New(),
		Status:    domain.InvoiceStatusOpen,
		AmountDue: 1000,
	}
	require.NoError(t, repo.Create(context.Background(), inv))

	paid, err := svc.Charge(context.Background(), inv.TenantID, inv.ID)
	require.NoError(t, err)
	require.Equal(t, domain.InvoiceStatusPaid, paid.Status)
}

func TestInvoiceService_MockChargeFailure(t *testing.T) {
	repo := &memoryInvoiceRepo{}
	cfg := &config.Config{BillingMockResult: "failure"}
	svc := NewInvoiceService(repo, nil, cfg, nil, nil, nil)
	inv := &domain.Invoice{
		TenantID:  uuid.New(),
		Status:    domain.InvoiceStatusOpen,
		AmountDue: 1000,
	}
	require.NoError(t, repo.Create(context.Background(), inv))

	_, err := svc.Charge(context.Background(), inv.TenantID, inv.ID)
	require.Error(t, err)
}
