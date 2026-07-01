package service

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const apiKeyPrefixLabel = "ssk"

func generateAPIKey() (fullKey, prefix string, hash string, err error) {
	prefixBytes := make([]byte, 4)
	secretBytes := make([]byte, 32)
	if _, err = rand.Read(prefixBytes); err != nil {
		return "", "", "", fmt.Errorf("generate prefix: %w", err)
	}
	if _, err = rand.Read(secretBytes); err != nil {
		return "", "", "", fmt.Errorf("generate secret: %w", err)
	}

	prefix = hex.EncodeToString(prefixBytes)
	secret := base64.RawURLEncoding.EncodeToString(secretBytes)
	fullKey = fmt.Sprintf("%s_%s_%s", apiKeyPrefixLabel, prefix, secret)

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(fullKey), bcrypt.DefaultCost)
	if err != nil {
		return "", "", "", fmt.Errorf("hash api key: %w", err)
	}

	return fullKey, prefix, string(hashBytes), nil
}

func generateWebhookSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate webhook secret: %w", err)
	}
	return hex.EncodeToString(b), nil
}
