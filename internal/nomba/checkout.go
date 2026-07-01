package nomba

import (
	"context"

	"github.com/ireoluwacodes/subsync/internal/domain"
)

// CreateOrder creates a checkout order for the given tenant.
func (c *Client) CreateOrder(ctx context.Context, tenant *domain.Tenant, req CreateOrderRequest) (CreateOrderResult, error) {
	return doData[CreateOrderResult](c, ctx, tenant, "POST", PathCheckoutOrder, req)
}
