package email

import "context"

type SendRequest struct {
	From    string
	To      string
	Subject string
	HTML    string
	Text    string
}

type EmailStrategy interface {
	Send(ctx context.Context, req SendRequest) error
}
