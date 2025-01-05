package utils

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sejo412/ya-metrics/internal/models"
)

// WithRetry wrapper for other functions with retries
func WithRetry(ctx context.Context, f func(ctx context.Context) error) error {
	var lastErr error
	for attempt := 0; attempt < models.RetryMaxRetries; attempt++ {
		err := f(ctx)
		if err == nil {
			return nil
		}
		if models.ErrIsRetryable(err) {
			lastErr = err
			delay := models.RetryInitDelay + time.Duration(attempt)*models.RetryDeltaDelay
			log.Printf("error: %v, retrying in %v", err, delay)
			time.Sleep(delay)
			continue
		}
		return err
	}
	return fmt.Errorf("error: All attempts failed [%d], last error is: %w", models.RetryMaxRetries, lastErr)
}
