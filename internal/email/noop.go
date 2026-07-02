package email

import (
	"context"

	"go.uber.org/zap"
)

type NoopStrategy struct{}

func (NoopStrategy) Send(ctx context.Context, req SendRequest) error {
	zap.L().Debug("email noop send",
		zap.String("to", req.To),
		zap.String("subject", req.Subject),
	)
	return nil
}
