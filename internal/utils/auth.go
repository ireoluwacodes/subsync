package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const apiKeyPrefixLabel = "ssk"

func HashResetSecret(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

func GenerateOTP() (string, error) {
	var n uint32
	if err := binary.Read(rand.Reader, binary.BigEndian, &n); err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n%1_000_000), nil
}

func GenerateResetToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

func GenerateAPIKey() (fullKey, prefix string, hash string, err error) {
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

func GenerateWebhookSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate webhook secret: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func GeneratePortalToken() (string, error) {
	return GenerateResetToken()
}
