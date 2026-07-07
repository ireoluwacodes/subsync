package observability

import (
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/ireoluwacodes/subsync/internal/config"
)

var sentryEnabled bool

// InitSentry configures the global Sentry client when SENTRY_DSN is set.
// Returns a flush function to call on shutdown.
func InitSentry(cfg *config.Config) (func(), error) {
	if cfg.SentryDSN == "" {
		sentryEnabled = false
		return func() {}, nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:         cfg.SentryDSN,
		Environment: cfg.AppEnv,
	})
	if err != nil {
		return nil, fmt.Errorf("sentry init: %w", err)
	}

	sentryEnabled = true
	return func() {
		sentry.Flush(2 * time.Second)
	}, nil
}

// CaptureJobError reports a background (asynq) job failure to Sentry.
func CaptureJobError(taskType string, err error, extras map[string]any) {
	if !sentryEnabled || err == nil {
		return
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelError)
		scope.SetTag("error.source", "worker_job")
		scope.SetTag("job.task", taskType)
		if len(extras) > 0 {
			ctx := make(sentry.Context, len(extras))
			for key, value := range extras {
				ctx[key] = value
			}
			scope.SetContext("worker_job", ctx)
		}
		scope.SetFingerprint([]string{"worker-job", taskType})
		sentry.CaptureException(err)
	})
}

// CaptureExternalAPIError reports outbound HTTP/SDK failures to Sentry.
func CaptureExternalAPIError(provider, operation string, err error, extras map[string]any) {
	if !sentryEnabled || err == nil {
		return
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelError)
		scope.SetTag("error.source", "external_api")
		scope.SetTag("external.provider", provider)
		if operation != "" {
			scope.SetTag("external.operation", operation)
		}
		if len(extras) > 0 {
			ctx := make(sentry.Context, len(extras))
			for key, value := range extras {
				ctx[key] = value
			}
			scope.SetContext("external_api", ctx)
		}
		scope.SetFingerprint([]string{"external-api", provider, operation})
		sentry.CaptureException(err)
	})
}
