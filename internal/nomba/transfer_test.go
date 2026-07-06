package nomba

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseBanksList_WrappedResults(t *testing.T) {
	raw := json.RawMessage(`{"results":[{"code":"058","name":"Guaranty Trust Bank"},{"code":"044","name":"Access Bank"}]}`)
	banks, err := parseBanksList(raw)
	require.NoError(t, err)
	require.Len(t, banks, 2)
	require.Equal(t, "058", banks[0].Code)
}

func TestParseBanksList_FlatArray(t *testing.T) {
	raw := json.RawMessage(`[{"code":"057","name":"Zenith Bank"}]`)
	banks, err := parseBanksList(raw)
	require.NoError(t, err)
	require.Len(t, banks, 1)
	require.Equal(t, "057", banks[0].Code)
}
