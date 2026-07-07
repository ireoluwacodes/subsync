package pdf

import (
	"strconv"
	"strings"

	"github.com/ireoluwacodes/subsync/internal/domain"
)

type brandTheme struct {
	CompanyName string
	AccentR     int
	AccentG     int
	AccentB     int
	MutedR      int
	MutedG      int
	MutedB      int
}

func themeFromTenant(t *domain.Tenant) brandTheme {
	theme := brandTheme{
		AccentR: 19,
		AccentG: 78,
		AccentB: 74,
		MutedR:  113,
		MutedG:  113,
		MutedB:  122,
	}
	if t == nil {
		return theme
	}
	theme.CompanyName = t.Name
	if name := brandingString(t.Branding, "company_name"); name != "" {
		theme.CompanyName = name
	}
	if hex := brandingString(t.Branding, "primary_color"); hex != "" {
		if r, g, b, ok := parseHexColor(hex); ok {
			theme.AccentR, theme.AccentG, theme.AccentB = r, g, b
		}
	}
	return theme
}

func brandingString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key].(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(v)
}

func parseHexColor(raw string) (r, g, b int, ok bool) {
	s := strings.TrimPrefix(strings.TrimSpace(raw), "#")
	if len(s) != 6 {
		return 0, 0, 0, false
	}
	rv, err1 := strconv.ParseInt(s[0:2], 16, 64)
	gv, err2 := strconv.ParseInt(s[2:4], 16, 64)
	bv, err3 := strconv.ParseInt(s[4:6], 16, 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, false
	}
	return int(rv), int(gv), int(bv), true
}
