package db

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/ireoluwacodes/subsync/internal/crypto"
	"github.com/ireoluwacodes/subsync/internal/db/models"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type TenantRepo struct {
	db  *DB
	enc *crypto.CredentialEncryptor
}

func NewTenantRepo(db *DB, enc *crypto.CredentialEncryptor) *TenantRepo {
	return &TenantRepo{db: db, enc: enc}
}

func parseAPIKeyPrefix(plaintextKey string) (string, error) {
	parts := strings.SplitN(plaintextKey, "_", 3)
	if len(parts) != 3 || parts[0] != "ssk" || parts[1] == "" || parts[2] == "" {
		return "", domain.ErrNotFound
	}
	return parts[1], nil
}

func (r *TenantRepo) encryptSecret(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	if r.enc == nil {
		return plain, nil
	}
	return r.enc.Encrypt(plain)
}

func (r *TenantRepo) decryptSecret(enc string) (string, error) {
	if enc == "" {
		return "", nil
	}
	if r.enc == nil {
		return enc, nil
	}
	return r.enc.Decrypt(enc)
}

func (r *TenantRepo) LoadNombaSecret(ctx context.Context, tenant *domain.Tenant) error {
	var m models.Tenant
	if err := r.db.WithContext(ctx).Select("nomba_client_secret_enc").First(&m, "id = ?", tenant.ID).Error; err != nil {
		return MapGORMError(err)
	}
	secret, err := r.decryptSecret(m.NombaClientSecretEnc)
	if err != nil {
		return err
	}
	tenant.NombaClientSecret = secret
	return nil
}

func (r *TenantRepo) LoadNombaWebhookSecret(ctx context.Context, tenant *domain.Tenant) error {
	var m models.Tenant
	if err := r.db.WithContext(ctx).Select("nomba_webhook_signing_key_enc").First(&m, "id = ?", tenant.ID).Error; err != nil {
		return MapGORMError(err)
	}
	secret, err := r.decryptSecret(m.NombaWebhookSigningKeyEnc)
	if err != nil {
		return err
	}
	tenant.NombaWebhookSecret = secret
	return nil
}

func (r *TenantRepo) resolveEncryptedSecrets(ctx context.Context, tenant *domain.Tenant) (clientEnc, webhookEnc string, err error) {
	var existing models.Tenant
	if tenant.ID != uuid.Nil {
		if err := r.db.WithContext(ctx).
			Select("nomba_client_secret_enc", "nomba_webhook_signing_key_enc").
			First(&existing, "id = ?", tenant.ID).Error; err != nil {
			return "", "", MapGORMError(err)
		}
	}

	if tenant.NombaClientSecret != "" {
		clientEnc, err = r.encryptSecret(tenant.NombaClientSecret)
		if err != nil {
			return "", "", err
		}
	} else {
		clientEnc = existing.NombaClientSecretEnc
	}

	if tenant.NombaWebhookSecret != "" {
		webhookEnc, err = r.encryptSecret(tenant.NombaWebhookSecret)
		if err != nil {
			return "", "", err
		}
	} else {
		webhookEnc = existing.NombaWebhookSigningKeyEnc
	}

	return clientEnc, webhookEnc, nil
}

func (r *TenantRepo) Create(ctx context.Context, tenant *domain.Tenant) error {
	clientEnc, err := r.encryptSecret(tenant.NombaClientSecret)
	if err != nil {
		return err
	}
	webhookEnc, err := r.encryptSecret(tenant.NombaWebhookSecret)
	if err != nil {
		return err
	}
	m, err := models.TenantFromDomain(tenant, clientEnc, webhookEnc)
	if err != nil {
		return err
	}

	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return MapGORMError(err)
	}

	*tenant = *models.TenantToDomain(m)
	tenant.NombaClientSecret = ""
	tenant.NombaWebhookSecret = ""
	return nil
}

func (r *TenantRepo) Update(ctx context.Context, tenant *domain.Tenant) error {
	clientEnc, webhookEnc, err := r.resolveEncryptedSecrets(ctx, tenant)
	if err != nil {
		return err
	}

	m, err := models.TenantFromDomain(tenant, clientEnc, webhookEnc)
	if err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Omit("CreatedAt").Save(m).Error; err != nil {
		return MapGORMError(err)
	}
	*tenant = *models.TenantToDomain(m)
	tenant.NombaClientSecret = ""
	tenant.NombaWebhookSecret = ""
	return nil
}

func (r *TenantRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	var m models.Tenant
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, MapGORMError(err)
	}
	return models.TenantToDomain(&m), nil
}

func (r *TenantRepo) AuthenticateAPIKey(ctx context.Context, plaintextKey string) (*domain.Tenant, error) {
	prefix, err := parseAPIKeyPrefix(plaintextKey)
	if err != nil {
		return nil, err
	}

	var m models.Tenant
	if err := r.db.WithContext(ctx).First(&m, "api_key_prefix = ?", prefix).Error; err != nil {
		return nil, MapGORMError(err)
	}

	tenant := models.TenantToDomain(&m)
	if err := bcrypt.CompareHashAndPassword([]byte(tenant.APIKeyHash), []byte(plaintextKey)); err != nil {
		return nil, domain.ErrNotFound
	}

	return tenant, nil
}

var _ domain.TenantRepository = (*TenantRepo)(nil)
