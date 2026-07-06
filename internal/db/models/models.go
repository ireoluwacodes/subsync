package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

type Tenant struct {
	ID                        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name                      string         `gorm:"not null"`
	Email                     string         `gorm:"not null;uniqueIndex"`
	Website                   string         `gorm:"not null;default:''"`
	NombaClientID             string         `gorm:"column:nomba_client_id;not null"`
	NombaClientSecretEnc      string         `gorm:"column:nomba_client_secret_enc;not null"`
	NombaAccountID            string         `gorm:"column:nomba_account_id;not null"`
	NombaSubAccountID         *string        `gorm:"column:nomba_sub_account_id"`
	NombaEnv                  string         `gorm:"column:nomba_env;not null;default:sandbox"`
	NombaWebhookSigningKeyEnc string         `gorm:"column:nomba_webhook_signing_key_enc;not null;default:''"`
	APIKeyPrefix              string         `gorm:"column:api_key_prefix;not null;index"`
	APIKeyHash                string         `gorm:"column:api_key_hash;not null"`
	WebhookSecret             string         `gorm:"column:webhook_secret;not null"`
	DunningConfig             datatypes.JSON `gorm:"column:dunning_config;type:jsonb;not null"`
	Branding                  datatypes.JSON `gorm:"column:branding;type:jsonb;not null;default:'{}'"`
	BillingEmail              datatypes.JSON `gorm:"column:billing_email;type:jsonb;not null;default:'{}'"`
	CreatedAt                 time.Time      `gorm:"not null"`
	UpdatedAt                 time.Time      `gorm:"not null"`
}

func (Tenant) TableName() string { return "tenants" }

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID     uuid.UUID `gorm:"type:uuid;not null;index"`
	Email        string    `gorm:"not null;uniqueIndex"`
	PasswordHash string    `gorm:"column:password_hash;not null"`
	Name         string    `gorm:"not null;default:''"`
	TokenVersion int       `gorm:"column:token_version;not null;default:1"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

func (User) TableName() string { return "users" }

type PasswordResetToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	TokenHash string     `gorm:"column:token_hash;not null"`
	ExpiresAt time.Time  `gorm:"not null"`
	UsedAt    *time.Time `gorm:"column:used_at"`
	CreatedAt time.Time  `gorm:"not null"`
}

func (PasswordResetToken) TableName() string { return "password_reset_tokens" }

type Subscription struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID           uuid.UUID  `gorm:"type:uuid;not null;index"`
	CustomerID         uuid.UUID  `gorm:"type:uuid;not null;index"`
	PlanID             uuid.UUID  `gorm:"type:uuid;not null"`
	PaymentMethodID           *uuid.UUID `gorm:"type:uuid"`
	FallbackPaymentMethodID   *uuid.UUID `gorm:"column:fallback_payment_method_id;type:uuid"`
	State                     string     `gorm:"not null;default:trialing"`
	TrialEndsAt        *time.Time `gorm:"column:trial_ends_at"`
	CurrentPeriodStart time.Time  `gorm:"column:current_period_start;not null"`
	CurrentPeriodEnd   time.Time  `gorm:"column:current_period_end;not null"`
	NextBillingAt      *time.Time `gorm:"column:next_billing_at"`
	CanceledAt         *time.Time `gorm:"column:canceled_at"`
	CancelAtPeriodEnd  bool       `gorm:"column:cancel_at_period_end;not null;default:false"`
	PauseStartsAt      *time.Time `gorm:"column:pause_starts_at"`
	PauseEndsAt        *time.Time `gorm:"column:pause_ends_at"`
	DunningStep        int        `gorm:"column:dunning_step;not null;default:0"`
	DunningStartedAt   *time.Time `gorm:"column:dunning_started_at"`
	Metadata           datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt          time.Time  `gorm:"not null"`
	UpdatedAt          time.Time  `gorm:"not null"`
}

func (Subscription) TableName() string { return "subscriptions" }

type SubscriptionTransition struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SubscriptionID uuid.UUID      `gorm:"type:uuid;not null;index"`
	TenantID       uuid.UUID      `gorm:"type:uuid;not null;index"`
	FromState      string         `gorm:"column:from_state;not null"`
	ToState        string         `gorm:"column:to_state;not null"`
	Reason         string         `gorm:"not null;default:''"`
	Actor          string         `gorm:"not null;default:''"`
	Metadata       datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt      time.Time      `gorm:"not null"`
}

func (SubscriptionTransition) TableName() string { return "subscription_transitions" }

type Invoice struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID           uuid.UUID      `gorm:"type:uuid;not null;index"`
	SubscriptionID     uuid.UUID      `gorm:"type:uuid;not null;index"`
	CustomerID         uuid.UUID      `gorm:"type:uuid;not null;index"`
	Status             string         `gorm:"not null;default:draft"`
	AmountDue          int64          `gorm:"column:amount_due;not null"`
	AmountPaid         int64          `gorm:"column:amount_paid;not null;default:0"`
	Currency           string         `gorm:"not null;default:NGN"`
	PeriodStart        time.Time      `gorm:"column:period_start;not null"`
	PeriodEnd          time.Time      `gorm:"column:period_end;not null"`
	DueDate            *time.Time     `gorm:"column:due_date"`
	PaidAt             *time.Time     `gorm:"column:paid_at"`
	VoidedAt           *time.Time     `gorm:"column:voided_at"`
	NombaOrderRef      *string        `gorm:"column:nomba_order_ref;uniqueIndex"`
	NombaTransactionID *string        `gorm:"column:nomba_transaction_id"`
	AttemptCount       int            `gorm:"column:attempt_count;not null;default:0"`
	NextAttemptAt      *time.Time     `gorm:"column:next_attempt_at"`
	Metadata           datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt          time.Time      `gorm:"not null"`
	UpdatedAt          time.Time      `gorm:"not null"`
}

func (Invoice) TableName() string { return "invoices" }

type InvoiceLineItem struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	InvoiceID   uuid.UUID  `gorm:"type:uuid;not null;index"`
	TenantID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	Type        string     `gorm:"not null"`
	Description string     `gorm:"not null"`
	Amount      int64      `gorm:"not null"`
	Currency    string     `gorm:"not null;default:NGN"`
	PeriodStart *time.Time `gorm:"column:period_start"`
	PeriodEnd   *time.Time `gorm:"column:period_end"`
	CreatedAt   time.Time  `gorm:"not null"`
}

func (InvoiceLineItem) TableName() string { return "invoice_line_items" }

type Plan struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID     uuid.UUID      `gorm:"type:uuid;not null;index"`
	Name         string         `gorm:"not null"`
	Description  *string        `gorm:"type:text"`
	Amount       int64          `gorm:"not null"`
	Currency     string         `gorm:"not null;default:NGN"`
	Interval     string         `gorm:"column:interval;not null"`
	IntervalDays *int           `gorm:"column:interval_days"`
	TrialDays    int            `gorm:"column:trial_days;not null;default:0"`
	Features     datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'"`
	IsActive     bool           `gorm:"column:is_active;not null;default:true"`
	IsArchived   bool           `gorm:"column:is_archived;not null;default:false"`
	CreatedAt    time.Time      `gorm:"not null"`
	UpdatedAt    time.Time      `gorm:"not null"`
}

func (Plan) TableName() string { return "plans" }

type Customer struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID   uuid.UUID      `gorm:"type:uuid;not null;index"`
	ExternalID *string        `gorm:"column:external_id"`
	Email      string         `gorm:"not null"`
	Name       *string        `gorm:"type:text"`
	Phone      *string        `gorm:"type:text"`
	Metadata   datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt  time.Time      `gorm:"not null"`
	UpdatedAt  time.Time      `gorm:"not null"`
}

func (Customer) TableName() string { return "customers" }

type PaymentMethod struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID   uuid.UUID `gorm:"type:uuid;not null;index"`
	CustomerID uuid.UUID `gorm:"type:uuid;not null;index"`
	Type       string    `gorm:"not null"`
	TokenKey   *string   `gorm:"column:token_key"`
	MandateID     *string `gorm:"column:mandate_id"`
	MandateStatus *string `gorm:"column:mandate_status"`
	CardLast4     *string `gorm:"column:card_last4"`
	CardBrand  *string   `gorm:"column:card_brand"`
	CardExpiry *string   `gorm:"column:card_expiry"`
	IsDefault  bool      `gorm:"column:is_default;not null;default:false"`
	CreatedAt  time.Time `gorm:"not null"`
	UpdatedAt  time.Time `gorm:"not null"`
}

func (PaymentMethod) TableName() string { return "payment_methods" }

type NombaEvent struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID      `gorm:"type:uuid;not null;index"`
	EventID     string         `gorm:"column:event_id;not null"`
	EventType   string         `gorm:"column:event_type;not null"`
	Payload     datatypes.JSON `gorm:"type:jsonb;not null"`
	Processed   bool           `gorm:"not null;default:false"`
	ProcessedAt *time.Time     `gorm:"column:processed_at"`
	Error       *string        `gorm:"type:text"`
	CreatedAt   time.Time      `gorm:"not null"`
}

func (NombaEvent) TableName() string { return "nomba_events" }

type WebhookEndpoint struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID  uuid.UUID      `gorm:"type:uuid;not null;index"`
	URL       string         `gorm:"not null"`
	Events    pq.StringArray `gorm:"type:text[];not null;default:'{}'"`
	IsActive  bool           `gorm:"column:is_active;not null;default:true"`
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
}

func (WebhookEndpoint) TableName() string { return "webhook_endpoints" }

type WebhookDelivery struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID     uuid.UUID  `gorm:"type:uuid;not null;index"`
	EndpointID   uuid.UUID  `gorm:"type:uuid;not null;index"`
	EventType    string     `gorm:"column:event_type;not null"`
	Payload      datatypes.JSON `gorm:"type:jsonb;not null"`
	AttemptCount int        `gorm:"column:attempt_count;not null;default:0"`
	LastStatus   *int       `gorm:"column:last_status"`
	LastError    *string    `gorm:"column:last_error;type:text"`
	DeliveredAt  *time.Time `gorm:"column:delivered_at"`
	NextRetryAt  *time.Time `gorm:"column:next_retry_at"`
	CreatedAt    time.Time  `gorm:"not null"`
}

func (WebhookDelivery) TableName() string { return "webhook_deliveries" }

type PortalToken struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	SubscriptionID uuid.UUID  `gorm:"type:uuid;not null"`
	CustomerID     uuid.UUID  `gorm:"type:uuid;not null"`
	TokenHash      string     `gorm:"column:token_hash;not null;uniqueIndex"`
	ExpiresAt      time.Time  `gorm:"not null"`
	UsedAt         *time.Time `gorm:"column:used_at"`
	CreatedAt      time.Time  `gorm:"not null"`
}

func (PortalToken) TableName() string { return "portal_tokens" }
