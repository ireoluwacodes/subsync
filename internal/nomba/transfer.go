package nomba

import (
	"context"
	"fmt"

	"github.com/ireoluwacodes/subsync/internal/domain"
)

// ListBanks returns supported Nigerian banks and their codes.
func (c *Client) ListBanks(ctx context.Context, tenant *domain.Tenant) ([]Bank, error) {
	return doData[[]Bank](c, ctx, tenant, "GET", PathTransfersBanks, nil)
}

// LookupBankAccount resolves the account holder name for an account number and bank code.
func (c *Client) LookupBankAccount(ctx context.Context, tenant *domain.Tenant, req BankAccountLookupRequest) (BankAccountLookupResult, error) {
	return doData[BankAccountLookupResult](c, ctx, tenant, "POST", PathTransfersBankLookup, req)
}

// BankTransfer sends funds from a sub-account to a bank account.
func (c *Client) BankTransfer(ctx context.Context, tenant *domain.Tenant, accountID string, req BankAccountTransferRequest) (BankAccountTransferResult, error) {
	path := fmt.Sprintf(PathSubAccountTransfer, accountID)
	return doData[BankAccountTransferResult](c, ctx, tenant, "POST", path, req)
}
