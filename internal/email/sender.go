package email

// Sender wraps the Resend transactional email client.

type Sender struct{}

func NewSender() *Sender { return &Sender{} }
