package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/julientant/weightlogr-cli/internal/logger"
)

func normalizeTimestamp(value string) (string, error) {
	if value == "" {
		return time.Now().UTC().Format(time.RFC3339), nil
	}

	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return "", fmt.Errorf("invalid timestamp %q (must be RFC3339): %w", value, err)
	}

	return t.UTC().Format(time.RFC3339), nil
}

func loadTimezone(tz string) (*time.Location, error) {
	if tz == "" {
		return time.UTC, nil
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %q: %w", tz, err)
	}
	return loc, nil
}

func withLogError(ctx context.Context, fn func() error) {
	if err := fn(); err != nil {
		logger.FromContext(ctx).Error("deferred close failed", "error", err)
	}
}
