package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeTimestamp(t *testing.T) {
	t.Run("empty string returns valid RFC3339 UTC timestamp", func(t *testing.T) {
		result, err := normalizeTimestamp("")
		require.NoError(t, err)
		assert.True(t, strings.HasSuffix(result, "Z"), "expected timestamp ending with Z, got %s", result)

		_, parseErr := time.Parse(time.RFC3339, result)
		assert.NoError(t, parseErr, "result should be valid RFC3339")
	})

	t.Run("valid UTC timestamp returns itself", func(t *testing.T) {
		input := "2026-04-05T15:00:00Z"
		result, err := normalizeTimestamp(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("valid offset timestamp is converted to UTC", func(t *testing.T) {
		result, err := normalizeTimestamp("2026-04-07T08:00:00-07:00")
		require.NoError(t, err)
		assert.Equal(t, "2026-04-07T15:00:00Z", result)
	})

	t.Run("invalid timestamp without T separator returns error", func(t *testing.T) {
		_, err := normalizeTimestamp("2026-04-05 08:00")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be RFC3339")
	})

	t.Run("completely invalid string returns error", func(t *testing.T) {
		_, err := normalizeTimestamp("not-a-date")
		require.Error(t, err)
	})
}

func TestLoadTimezone(t *testing.T) {
	t.Run("empty string returns UTC", func(t *testing.T) {
		loc, err := loadTimezone("")
		require.NoError(t, err)
		assert.Equal(t, time.UTC, loc)
	})

	t.Run("UTC returns UTC", func(t *testing.T) {
		loc, err := loadTimezone("UTC")
		require.NoError(t, err)
		assert.Equal(t, time.UTC, loc)
	})

	t.Run("valid IANA timezone returns location", func(t *testing.T) {
		loc, err := loadTimezone("America/Phoenix")
		require.NoError(t, err)
		assert.NotNil(t, loc)
	})

	t.Run("invalid timezone returns error", func(t *testing.T) {
		_, err := loadTimezone("Invalid/Zone")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid timezone")
	})
}
