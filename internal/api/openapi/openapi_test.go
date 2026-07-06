package openapi

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocument_IntegratorRoutesOnly(t *testing.T) {
	var spec struct {
		Paths map[string]any `json:"paths"`
		Components struct {
			SecuritySchemes map[string]any `json:"securitySchemes"`
		} `json:"components"`
	}
	require.NoError(t, json.Unmarshal(Document, &spec))

	excluded := []string{
		"/auth/register",
		"/auth/login",
		"/auth/refresh",
		"/auth/logout",
		"/me",
		"/settings",
		"/analytics/mrr",
		"/analytics/churn",
		"/analytics/dunning",
		"/analytics/revenue",
	}
	for _, path := range excluded {
		_, ok := spec.Paths[path]
		require.False(t, ok, "dashboard route should not be in openapi: %s", path)
	}

	require.Contains(t, spec.Paths, "/subscriptions/checkout")
	require.Contains(t, spec.Paths, "/portal/token")
	require.Contains(t, spec.Components.SecuritySchemes, "apiKeyAuth")
}

func TestDocument_DocumentsMinorUnitMoney(t *testing.T) {
	var spec struct {
		Info struct {
			Description string `json:"description"`
		} `json:"info"`
		Components struct {
			Schemas map[string]struct {
				Description string `json:"description"`
			} `json:"schemas"`
		} `json:"components"`
	}
	require.NoError(t, json.Unmarshal(Document, &spec))
	require.Contains(t, spec.Info.Description, "minor unit")
	require.Contains(t, spec.Info.Description, "kobo")
	require.Contains(t, spec.Components.Schemas["MoneyMinorAmount"].Description, "kobo")
}

func TestDocument_EveryOperationHasResponses(t *testing.T) {
	var spec struct {
		Paths map[string]map[string]struct {
			Responses map[string]any `json:"responses"`
		} `json:"paths"`
	}
	require.NoError(t, json.Unmarshal(Document, &spec))

	for path, item := range spec.Paths {
		for method, op := range item {
			if strings.HasPrefix(method, "/") {
				continue
			}
			require.NotEmpty(t, op.Responses, "%s %s must define responses", method, path)
		}
	}
}
