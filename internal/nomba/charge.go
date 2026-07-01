package nomba

import (
	"context"

	"github.com/ireoluwacodes/subsync/internal/domain"
)

// TokenizedCardPayment charges a stored card token for the given tenant.
func (c *Client) TokenizedCardPayment(ctx context.Context, tenant *domain.Tenant, req TokenizedCardPaymentRequest) (TokenizedCardPaymentResult, error) {
	return doData[TokenizedCardPaymentResult](c, ctx, tenant, "POST", PathTokenizedPayment, req)
}
