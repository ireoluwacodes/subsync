package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/ireoluwacodes/subsync/internal/api/dto"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

const (
	TenantContextKey   = "tenant"
	TenantIDContextKey = "tenant_id"
	UserContextKey     = "user"
	UserIDContextKey   = "user_id"
	authCachePrefix    = "auth:tenant:"
	authCacheTTL       = 5 * time.Minute
)

type TenantAuthenticator interface {
	AuthenticateAPIKey(ctx context.Context, plaintextKey string) (*domain.Tenant, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
}

type JWTValidator interface {
	ValidateAccessToken(ctx context.Context, token string) (*domain.User, *domain.Tenant, error)
}

func Auth(tenants TenantAuthenticator, jwt JWTValidator, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			abortUnauthorized(c)
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if token == "" {
			abortUnauthorized(c)
			return
		}

		if strings.HasPrefix(token, "ssk_") {
			tenant, err := resolveTenantByAPIKey(c.Request.Context(), tenants, rdb, token)
			if err != nil {
				abortUnauthorized(c)
				return
			}
			c.Set(TenantContextKey, tenant)
			c.Set(TenantIDContextKey, tenant.ID)
			c.Next()
			return
		}

		if jwt == nil {
			abortUnauthorized(c)
			return
		}

		user, tenant, err := jwt.ValidateAccessToken(c.Request.Context(), token)
		if err != nil {
			abortUnauthorized(c)
			return
		}

		c.Set(TenantContextKey, tenant)
		c.Set(TenantIDContextKey, tenant.ID)
		c.Set(UserContextKey, user)
		c.Set(UserIDContextKey, user.ID)
		c.Next()
	}
}

func resolveTenantByAPIKey(ctx context.Context, tenants TenantAuthenticator, rdb *redis.Client, apiKey string) (*domain.Tenant, error) {
	if rdb != nil {
		cacheKey := authCachePrefix + sha256Hex(apiKey)
		if tenantID, err := rdb.Get(ctx, cacheKey).Result(); err == nil {
			if id, parseErr := uuid.Parse(tenantID); parseErr == nil {
				if tenant, err := tenants.GetByID(ctx, id); err == nil {
					return tenant, nil
				}
			}
		}
	}

	tenant, err := tenants.AuthenticateAPIKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	if rdb != nil {
		cacheKey := authCachePrefix + sha256Hex(apiKey)
		_ = rdb.Set(ctx, cacheKey, tenant.ID.String(), authCacheTTL).Err()
	}

	return tenant, nil
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func abortUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Envelope{
		Meta: dto.Meta{RequestID: c.GetString("request_id")},
		Error: &dto.APIError{
			Code:    "unauthorized",
			Message: "missing or invalid credentials",
		},
	})
}

func TenantFromContext(c *gin.Context) (*domain.Tenant, bool) {
	val, ok := c.Get(TenantContextKey)
	if !ok {
		return nil, false
	}
	tenant, ok := val.(*domain.Tenant)
	return tenant, ok && tenant != nil
}

func UserFromContext(c *gin.Context) (*domain.User, bool) {
	val, ok := c.Get(UserContextKey)
	if !ok {
		return nil, false
	}
	user, ok := val.(*domain.User)
	return user, ok && user != nil
}

func TenantIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	val, ok := c.Get(TenantIDContextKey)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

func Tenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := TenantFromContext(c); !ok {
			abortUnauthorized(c)
			return
		}
		c.Next()
	}
}
