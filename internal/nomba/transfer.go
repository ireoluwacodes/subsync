package nomba

import (
	"context"
	"fmt"

	"github.com/ireoluwacodes/subsync/internal/domain"
)

// BankTransfer sends funds from a sub-account to a bank account.
func (c *Client) BankTransfer(ctx context.Context, tenant *domain.Tenant, accountID string, req BankAccountTransferRequest) (BankAccountTransferResult, error) {
	path := fmt.Sprintf(PathSubAccountTransfer, accountID)
	return doData[BankAccountTransferResult](c, ctx, tenant, "POST", path, req)
}
