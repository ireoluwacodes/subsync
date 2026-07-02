package email

import (
	"fmt"
	"html"
	"strings"
)

const (
	colorBg      = "#F7F6F3"
	colorSurface = "#FFFFFF"
	colorText    = "#18181B"
	colorMuted   = "#71717A"
	colorBorder  = "#E4E4E7"
	colorAccent  = "#134E4A"
	colorWarn    = "#B45309"
	colorDanger  = "#991B1B"
)

func formatAmount(amount int64, currency string) string {
	major := float64(amount) / 100
	switch strings.ToUpper(currency) {
	case "NGN":
		return fmt.Sprintf("₦%s", formatMajor(major))
	case "USD":
		return fmt.Sprintf("$%s", formatMajor(major))
	case "GBP":
		return fmt.Sprintf("£%s", formatMajor(major))
	case "EUR":
		return fmt.Sprintf("€%s", formatMajor(major))
	default:
		return fmt.Sprintf("%s %s", strings.ToUpper(currency), formatMajor(major))
	}
}

func formatMajor(v float64) string {
	if v == float64(int64(v)) {
		return fmt.Sprintf("%d", int64(v))
	}
	return fmt.Sprintf("%.2f", v)
}

type layoutOpts struct {
	eyebrow   string
	heading   string
	body      string
	ctaLabel  string
	ctaURL    string
	footer    string
	accent    string
}

func emailLayout(opts layoutOpts) string {
	accent := opts.accent
	if accent == "" {
		accent = colorAccent
	}

	eyebrow := ""
	if opts.eyebrow != "" {
		eyebrow = fmt.Sprintf(
			`<p style="margin:0 0 12px;font-size:11px;font-weight:600;letter-spacing:0.12em;text-transform:uppercase;color:%s;">%s</p>`,
			accent, html.EscapeString(opts.eyebrow),
		)
	}

	cta := ""
	if opts.ctaLabel != "" && opts.ctaURL != "" {
		cta = fmt.Sprintf(`
<tr>
  <td style="padding:28px 0 0;">
    <a href="%s" style="display:inline-block;padding:14px 28px;background:%s;color:#FFFFFF;font-size:14px;font-weight:600;text-decoration:none;border-radius:6px;letter-spacing:0.01em;">%s</a>
  </td>
</tr>`, html.EscapeString(opts.ctaURL), accent, html.EscapeString(opts.ctaLabel))
	}

	footer := "You're receiving this because you have an account with SubSync."
	if opts.footer != "" {
		footer = opts.footer
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s</title>
</head>
<body style="margin:0;padding:0;background:%s;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;-webkit-font-smoothing:antialiased;">
  <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background:%s;">
    <tr>
      <td align="center" style="padding:48px 20px;">
        <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="max-width:520px;">
          <tr>
            <td style="padding:0 0 32px;text-align:center;">
              <span style="font-size:13px;font-weight:600;letter-spacing:0.18em;text-transform:uppercase;color:%s;">SubSync</span>
            </td>
          </tr>
          <tr>
            <td style="background:%s;border:1px solid %s;border-radius:10px;padding:40px 36px;">
              <table role="presentation" width="100%%" cellpadding="0" cellspacing="0">
                <tr>
                  <td>
                    %s
                    <h1 style="margin:0 0 16px;font-size:22px;font-weight:600;line-height:1.35;color:%s;letter-spacing:-0.02em;">%s</h1>
                    <div style="font-size:15px;line-height:1.65;color:%s;">%s</div>
                  </td>
                </tr>
                %s
              </table>
            </td>
          </tr>
          <tr>
            <td style="padding:28px 8px 0;text-align:center;font-size:12px;line-height:1.6;color:%s;">%s</td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`,
		html.EscapeString(opts.heading),
		colorBg, colorBg,
		colorMuted,
		colorSurface, colorBorder,
		eyebrow,
		colorText, html.EscapeString(opts.heading),
		colorMuted, opts.body,
		cta,
		colorMuted, html.EscapeString(footer),
	)
}

func amountLine(amount int64, currency string) string {
	if amount <= 0 {
		return ""
	}
	return fmt.Sprintf(
		`<p style="margin:20px 0 0;padding:16px 18px;background:%s;border-radius:8px;font-size:20px;font-weight:600;color:%s;letter-spacing:-0.02em;">%s</p>`,
		colorBg, colorText, html.EscapeString(formatAmount(amount, currency)),
	)
}

func PaymentFailedHTML(tenantName string, amount int64, currency string) (subject, htmlOut string) {
	subject = "We couldn't process your payment"
	tenant := html.EscapeString(tenantName)
	body := fmt.Sprintf(
		`<p style="margin:0;">We were unable to charge your payment method for <strong style="color:%s;">%s</strong>.</p>
		<p style="margin:16px 0 0;">Please update your billing details to keep your subscription active.</p>%s`,
		colorText, tenant, amountLine(amount, currency),
	)
	htmlOut = emailLayout(layoutOpts{
		eyebrow: "Payment failed",
		heading: "Action needed",
		body:    body,
		accent:  colorDanger,
	})
	return subject, htmlOut
}

func DunningWarningHTML(tenantName string, step int) (subject, htmlOut string) {
	subject = "Reminder: update your payment"
	tenant := html.EscapeString(tenantName)
	body := fmt.Sprintf(
		`<p style="margin:0;">This is reminder <strong style="color:%s;">%d</strong> about your overdue subscription with <strong style="color:%s;">%s</strong>.</p>
		<p style="margin:16px 0 0;">Update your payment method to avoid interruption.</p>`,
		colorText, step, colorText, tenant,
	)
	htmlOut = emailLayout(layoutOpts{
		eyebrow: fmt.Sprintf("Reminder %d", step),
		heading: "Your payment is overdue",
		body:    body,
		accent:  colorWarn,
	})
	return subject, htmlOut
}

func DunningFinalHTML(tenantName string) (subject, htmlOut string) {
	subject = "Your subscription has been canceled"
	tenant := html.EscapeString(tenantName)
	body := fmt.Sprintf(
		`<p style="margin:0;">After several unsuccessful payment attempts, your subscription with <strong style="color:%s;">%s</strong> has been canceled.</p>
		<p style="margin:16px 0 0;">You can resubscribe at any time when you're ready.</p>`,
		colorText, tenant,
	)
	htmlOut = emailLayout(layoutOpts{
		eyebrow: "Subscription ended",
		heading: "Subscription canceled",
		body:    body,
		accent:  colorMuted,
	})
	return subject, htmlOut
}

func SubscriptionConfirmedHTML(tenantName string, amount int64, currency string) (subject, htmlOut string) {
	subject = "Payment received — thank you"
	tenant := html.EscapeString(tenantName)
	body := fmt.Sprintf(
		`<p style="margin:0;">We've received your payment for <strong style="color:%s;">%s</strong>. Your subscription is active.</p>%s`,
		colorText, tenant, amountLine(amount, currency),
	)
	htmlOut = emailLayout(layoutOpts{
		eyebrow: "Payment confirmed",
		heading: "Thank you",
		body:    body,
	})
	return subject, htmlOut
}

func PasswordResetOTPHTML(otp string) (subject, htmlOut string) {
	subject = "Your SubSync password reset code"
	body := fmt.Sprintf(
		`<p style="margin:0;">Use this one-time code to reset your SubSync password. It expires in 10 minutes.</p>
		<p style="margin:20px 0 0;padding:16px 18px;background:%s;border-radius:8px;font-size:28px;font-weight:600;letter-spacing:0.28em;text-align:center;color:%s;">%s</p>
		<p style="margin:16px 0 0;">If you didn't request this, you can safely ignore this email.</p>`,
		colorBg, colorText, html.EscapeString(otp),
	)
	htmlOut = emailLayout(layoutOpts{
		eyebrow: "Password reset",
		heading: "Your verification code",
		body:    body,
	})
	return subject, htmlOut
}
