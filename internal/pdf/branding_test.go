package pdf

import (
	"testing"

	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestThemeFromTenant_usesBranding(t *testing.T) {
	theme := themeFromTenant(&domain.Tenant{
		Name: "Acme",
		Branding: map[string]any{
			"company_name":  "Acme Ltd",
			"primary_color": "#FF00AA",
		},
	})
	require.Equal(t, "Acme Ltd", theme.CompanyName)
	require.Equal(t, 255, theme.AccentR)
	require.Equal(t, 0, theme.AccentG)
	require.Equal(t, 170, theme.AccentB)
}
