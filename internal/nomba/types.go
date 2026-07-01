package nomba

// Nomba API request/response types are split by domain:
//   - types_envelope.go — APIResponse, APIError
//   - types_auth.go — token issue/refresh/revoke
//   - types_checkout.go — orders, tokenized card payment
//   - types_direct_debit.go — mandates
//   - types_transfer.go — bank transfers, balance lookup
//   - types_webhook.go — inbound webhook payloads
