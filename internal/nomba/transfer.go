package nomba

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ireoluwacodes/subsync/internal/domain"
)

// ListBanks returns Nigerian banks from Nomba (all NIP participants).
func (c *Client) ListBanks(ctx context.Context, tenant *domain.Tenant) ([]Bank, error) {
	var resp APIResponse[json.RawMessage]
	if err := c.do(ctx, tenant, "GET", PathTransfersBanks, nil, &resp); err != nil {
		return nil, err
	}
	if !resp.OK() {
		return nil, NewHTTPError(http.StatusOK, APIError{Code: resp.Code, Description: resp.Description})
	}
	return parseBanksList(resp.Data)
}

// ListDirectDebitBanks returns banks that support NIBSS e-mandate mandates.
func (c *Client) ListDirectDebitBanks(ctx context.Context, tenant *domain.Tenant) ([]Bank, error) {
	banks, err := c.ListBanks(ctx, tenant)
	if err != nil {
		return nil, err
	}
	return FilterBanksForDirectDebit(banks), nil
}

func parseBanksList(raw json.RawMessage) ([]Bank, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var wrapped BanksListResults
	if err := json.Unmarshal(raw, &wrapped); err == nil && len(wrapped.Results) > 0 {
		return wrapped.Results, nil
	}
	var flat []Bank
	if err := json.Unmarshal(raw, &flat); err != nil {
		return nil, fmt.Errorf("decode banks list: %w", err)
	}
	return flat, nil
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
