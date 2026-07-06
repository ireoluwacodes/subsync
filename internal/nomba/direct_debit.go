package nomba

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/ireoluwacodes/subsync/internal/domain"
	"go.uber.org/zap"
)

// CreateMandate registers a direct debit mandate for the tenant.
func (c *Client) CreateMandate(ctx context.Context, tenant *domain.Tenant, req CreateMandateRequest) (CreateMandateResult, error) {
	var raw json.RawMessage
	if err := c.do(ctx, tenant, "POST", PathDirectDebitCreate, req, &raw); err != nil {
		return CreateMandateResult{}, err
	}
	result, err := parseMandateCreateResponse(raw)
	if err != nil {
		c.logNombaResponseFailure(tenant, "POST", PathDirectDebitCreate, raw, err)
		return CreateMandateResult{}, err
	}
	return result, nil
}

// GetMandateStatus returns the current mandate status.
func (c *Client) GetMandateStatus(ctx context.Context, tenant *domain.Tenant, mandateID string) (MandateStatusResult, error) {
	path := PathDirectDebitStatus + "?mandateId=" + url.QueryEscape(mandateID)
	var raw json.RawMessage
	if err := c.do(ctx, tenant, "GET", path, nil, &raw); err != nil {
		return MandateStatusResult{}, err
	}
	result, err := parseMandateStatusResponse(raw)
	if err != nil {
		c.logNombaResponseFailure(tenant, "GET", path, raw, err)
		return MandateStatusResult{}, err
	}
	return result, nil
}

// DebitMandate debits an active mandate.
func (c *Client) DebitMandate(ctx context.Context, tenant *domain.Tenant, req DebitMandateRequest) (DebitMandateResult, error) {
	return doData[DebitMandateResult](c, ctx, tenant, "POST", PathDirectDebitDebit, req)
}

func (c *Client) logNombaResponseFailure(tenant *domain.Tenant, method, path string, raw json.RawMessage, err error) {
	if c.log == nil {
		return
	}
	fields := []zap.Field{
		zap.String("method", method),
		zap.String("path", path),
		zap.String("response_body", string(raw)),
		zap.Error(err),
	}
	if tenant != nil {
		fields = append(fields, zap.String("tenant_id", tenant.ID.String()))
	}
	c.log.Warn("nomba api business error", fields...)
}
