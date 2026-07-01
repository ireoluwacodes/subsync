package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

var ErrInvalidKey = errors.New("encryption key must be 32 bytes")

type CredentialEncryptor struct {
	aead cipher.AEAD
}

func NewCredentialEncryptor(key []byte) (*CredentialEncryptor, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}
	return &CredentialEncryptor{aead: aead}, nil
}

func ParseKey(raw string) ([]byte, error) {
	if raw == "" {
		return nil, ErrInvalidKey
	}
	if b, err := base64.StdEncoding.DecodeString(raw); err == nil && len(b) == 32 {
		return b, nil
	}
	if len(raw) == 64 {
		out := make([]byte, 32)
		for i := 0; i < 32; i++ {
			var v byte
			_, err := fmt.Sscanf(raw[i*2:i*2+2], "%02x", &v)
			if err != nil {
				return nil, ErrInvalidKey
			}
			out[i] = v
		}
		return out, nil
	}
	if len(raw) == 32 {
		return []byte(raw), nil
	}
	return nil, ErrInvalidKey
}

func (e *CredentialEncryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	nonce := make([]byte, e.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := e.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *CredentialEncryptor) Decrypt(encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode ciphertext: %w", err)
	}
	nonceSize := e.aead.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plain, err := e.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(plain), nil
}
