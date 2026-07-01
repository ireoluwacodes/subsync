package crypto

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCredentialEncryptorRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	enc, err := NewCredentialEncryptor(key)
	require.NoError(t, err)

	ciphertext, err := enc.Encrypt("super-secret-nomba-key")
	require.NoError(t, err)
	require.NotEqual(t, "super-secret-nomba-key", ciphertext)

	plain, err := enc.Decrypt(ciphertext)
	require.NoError(t, err)
	require.Equal(t, "super-secret-nomba-key", plain)
}

func TestParseKeyBase64(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	raw := base64.StdEncoding.EncodeToString(key)
	imported, err := ParseKey(raw)
	require.NoError(t, err)
	require.Equal(t, key, imported)
}
