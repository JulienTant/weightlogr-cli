package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/viper"

	"github.com/julientant/weightlogr-cli/internal/logger"
)

const (
	TimeFormatDateTimeMinute = "2006-01-02T15:04"
	TimeFormatDateTimeSecond = "2006-01-02T15:04:05"
	TimeFormatStorage        = "2006-01-02 15:04:05"

	LogFilePermissions = 0o644
)

func loadTimezone(ctx context.Context) (*time.Location, error) {
	tz, err := time.LoadLocation(viper.GetString("timezone"))
	if err != nil {
		return nil, fmt.Errorf("invalid timezone: %w", err)
	}
	logger.FromContext(ctx).DebugContext(ctx, "loaded timezone", "tz", tz.String())
	return tz, nil
}

func normalizeTimestamp(ctx context.Context, value string, tz *time.Location) (string, error) {
	if value == "" {
		return time.Now().In(tz).Format(TimeFormatStorage), nil
	}

	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t, err = time.ParseInLocation(TimeFormatDateTimeMinute, value, tz)
		if err != nil {
			t, err = time.ParseInLocation(TimeFormatDateTimeSecond, value, tz)
			if err != nil {
				return "", fmt.Errorf("invalid timestamp %q: %w", value, err)
			}
		}
	}

	return t.In(tz).Format(TimeFormatStorage), nil
}

func withLogError(ctx context.Context, fn func() error) {
	if err := fn(); err != nil {
		logger.FromContext(ctx).Error("deferred close failed", "error", err)
	}
}
