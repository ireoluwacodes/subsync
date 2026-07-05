package nomba

import (
	"context"
	"net/http"
	"net/url"

	"github.com/ireoluwacodes/subsync/internal/domain"
)

// CreateMandate registers a direct debit mandate for the tenant.
func (c *Client) CreateMandate(ctx context.Context, tenant *domain.Tenant, req CreateMandateRequest) (CreateMandateResult, error) {
	var resp LegacyMandateAPIResponse
	if err := c.do(ctx, tenant, "POST", PathDirectDebitCreate, req, &resp); err != nil {
		return CreateMandateResult{}, err
	}
	if !resp.OK() {
		return CreateMandateResult{}, NewHTTPError(http.StatusOK, APIError{
			Code:        resp.ResponseCode,
			Description: resp.ResponseMessage,
		})
	}
	return resp.Data, nil
}

// GetMandateStatus returns the current mandate status.
func (c *Client) GetMandateStatus(ctx context.Context, tenant *domain.Tenant, mandateID string) (MandateStatusResult, error) {
	path := PathDirectDebitStatus + "?mandateId=" + url.QueryEscape(mandateID)
	return doData[MandateStatusResult](c, ctx, tenant, "GET", path, nil)
}

// DebitMandate debits an active mandate.
func (c *Client) DebitMandate(ctx context.Context, tenant *domain.Tenant, req DebitMandateRequest) (DebitMandateResult, error) {
	return doData[DebitMandateResult](c, ctx, tenant, "POST", PathDirectDebitDebit, req)
}
