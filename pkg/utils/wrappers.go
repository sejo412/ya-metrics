package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/sejo412/ya-metrics/internal/logger"
	"github.com/sejo412/ya-metrics/internal/models"
)

// WithRetry wrapper for other functions with retries.
func WithRetry(ctx context.Context, log *logger.Logger, f func(ctx context.Context) error) error {
	var lastErr error
	for attempt := 0; attempt < models.RetryMaxRetries; attempt++ {
		err := f(ctx)
		if err == nil {
			return nil
		}
		if models.ErrIsRetryable(err) {
			lastErr = err
			delay := models.RetryInitDelay + time.Duration(attempt)*models.RetryDeltaDelay
			log.Logger.Errorw("attempt failed",
				"attempt", attempt,
				"delay", delay,
				"error", lastErr)
			time.Sleep(delay)
			continue
		}
		return err
	}
	return fmt.Errorf("all attempts failed [%d], last error is: %w", models.RetryMaxRetries, lastErr)
}
