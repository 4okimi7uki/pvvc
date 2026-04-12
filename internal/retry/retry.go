package retry

import (
	"context"
	"fmt"
	"time"
)

func Do(ctx context.Context, maxAttempts int, fn func() error) error {
	var err error
	for i := range maxAttempts {
		err = fn()
		if err == nil {
			return nil
		}
		if i < maxAttempts-1 {
			wait := time.Duration(1<<i) * time.Second
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return fmt.Errorf("after %d attempts: %w", maxAttempts, err)
}
