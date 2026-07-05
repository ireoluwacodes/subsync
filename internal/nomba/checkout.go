package nomba

import (
	"context"
	"net/url"

	"github.com/ireoluwacodes/subsync/internal/domain"
)

// CreateOrder creates a checkout order for the given tenant.
func (c *Client) CreateOrder(ctx context.Context, tenant *domain.Tenant, req CreateOrderRequest) (CreateOrderResult, error) {
	return doData[CreateOrderResult](c, ctx, tenant, "POST", PathCheckoutOrder, req)
}

// VerifyCheckoutTransaction looks up a checkout transaction by order reference.
func (c *Client) VerifyCheckoutTransaction(ctx context.Context, tenant *domain.Tenant, orderReference string) (CheckoutTransactionDetailsResult, error) {
	path := PathCheckoutVerify + "?orderReference=" + url.QueryEscape(orderReference)
	return doData[CheckoutTransactionDetailsResult](c, ctx, tenant, "GET", path, nil)
}
