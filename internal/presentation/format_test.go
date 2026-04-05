package presentation

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/julientant/weightlogr-cli/internal/store"
)

func sampleWeighIn() store.WeighIn {
	return store.WeighIn{
		ID:        1,
		Weight:    185.5,
		CreatedAt: "2026-04-05 08:30:00",
		Source:    "manual",
		Notes:     "morning",
	}
}

func TestFormatInsert(t *testing.T) {
	t.Run("table", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, FormatInsert(&buf, "table", "kg", sampleWeighIn()))

		out := buf.String()
		assert.Contains(t, out, "Logged")
		assert.Contains(t, out, "185.5")
		assert.Contains(t, out, "kg")
		assert.Contains(t, out, "2026-04-05 08:30:00")
		assert.Contains(t, out, "row id 1")
	})

	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		r := sampleWeighIn()
		require.NoError(t, FormatInsert(&buf, "json", "kg", r))

		var decoded store.WeighIn
		require.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))

		assert.Equal(t, r.ID, decoded.ID)
		assert.InDelta(t, r.Weight, decoded.Weight, 0.001)
		assert.Equal(t, r.CreatedAt, decoded.CreatedAt)
		assert.Equal(t, r.Source, decoded.Source)
		assert.Equal(t, r.Notes, decoded.Notes)
	})

	t.Run("csv", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, FormatInsert(&buf, "csv", "kg", sampleWeighIn()))

		records, err := csv.NewReader(strings.NewReader(buf.String())).ReadAll()
		require.NoError(t, err)

		require.Len(t, records, 2)
		assert.Equal(t, []string{"id", "weight", "created_at", "source", "notes"}, records[0])
		assert.Equal(t, "1", records[1][0])
		assert.Equal(t, "185.5", records[1][1])
		assert.Equal(t, "2026-04-05 08:30:00", records[1][2])
		assert.Equal(t, "manual", records[1][3])
		assert.Equal(t, "morning", records[1][4])
	})

	t.Run("csv notes with comma", func(t *testing.T) {
		r := sampleWeighIn()
		r.Notes = "felt tired, sore"

		var buf bytes.Buffer
		require.NoError(t, FormatInsert(&buf, "csv", "kg", r))

		records, err := csv.NewReader(strings.NewReader(buf.String())).ReadAll()
		require.NoError(t, err)

		require.Len(t, records, 2)
		assert.Equal(t, "felt tired, sore", records[1][4])
	})

	t.Run("unknown format defaults to table", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, FormatInsert(&buf, "xml", "kg", sampleWeighIn()))

		out := buf.String()
		assert.Contains(t, out, "Logged")
		assert.Contains(t, out, "185.5")
	})
}

func TestFormatList(t *testing.T) {
	twoEntries := []store.WeighIn{
		{ID: 1, Weight: 185.0, CreatedAt: "2026-04-01 08:00:00", Source: "manual", Notes: "good day"},
		{ID: 2, Weight: 184.5, CreatedAt: "2026-04-02 08:00:00", Source: "scale", Notes: ""},
	}

	t.Run("table empty", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, FormatList(&buf, "table", "kg", nil))
		assert.Equal(t, "No weigh-ins found.\n", buf.String())
	})

	t.Run("table with entries", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, FormatList(&buf, "table", "kg", twoEntries))

		lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
		require.GreaterOrEqual(t, len(lines), 4)

		assert.Contains(t, lines[0], "Timestamp")
		assert.Contains(t, lines[0], "Weight")
		assert.Contains(t, lines[1], "---")
		assert.Contains(t, lines[2], "185.0")
		assert.Contains(t, lines[2], "good day")
		assert.Contains(t, lines[3], "184.5")
		assert.Contains(t, lines[3], "| -", "empty notes should render as dash")
	})

	t.Run("json empty", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, FormatList(&buf, "json", "kg", nil))
		assert.Equal(t, "null\n", buf.String())
	})

	t.Run("json with entries", func(t *testing.T) {
		entries := []store.WeighIn{
			{ID: 1, Weight: 185.0, CreatedAt: "2026-04-01 08:00:00", Source: "manual", Notes: "a"},
			{ID: 2, Weight: 184.5, CreatedAt: "2026-04-02 08:00:00", Source: "scale", Notes: "b"},
		}

		var buf bytes.Buffer
		require.NoError(t, FormatList(&buf, "json", "kg", entries))

		var decoded []store.WeighIn
		require.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))

		require.Len(t, decoded, 2)
		assert.Equal(t, int64(1), decoded[0].ID)
		assert.InDelta(t, 185.0, decoded[0].Weight, 0.001)
		assert.Equal(t, "a", decoded[0].Notes)
	})

	t.Run("csv", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, FormatList(&buf, "csv", "kg", twoEntries))

		records, err := csv.NewReader(strings.NewReader(buf.String())).ReadAll()
		require.NoError(t, err)

		require.Len(t, records, 3)
		assert.Equal(t, []string{"id", "weight", "created_at", "source", "notes"}, records[0])
		assert.Equal(t, "185.0", records[1][1])
		assert.Equal(t, "184.5", records[2][1])
	})

	t.Run("csv special characters in notes", func(t *testing.T) {
		entries := []store.WeighIn{
			{ID: 1, Weight: 185.0, CreatedAt: "2026-04-01 08:00:00", Source: "manual", Notes: "tired, sore"},
			{ID: 2, Weight: 184.5, CreatedAt: "2026-04-02 08:00:00", Source: "scale", Notes: `she said "wow"`},
		}

		var buf bytes.Buffer
		require.NoError(t, FormatList(&buf, "csv", "kg", entries))

		records, err := csv.NewReader(strings.NewReader(buf.String())).ReadAll()
		require.NoError(t, err)

		require.Len(t, records, 3)
		assert.Equal(t, "tired, sore", records[1][4])
		assert.Equal(t, `she said "wow"`, records[2][4])
	})

	t.Run("multiple entries all present", func(t *testing.T) {
		var entries []store.WeighIn
		for i := 0; i < 5; i++ {
			entries = append(entries, store.WeighIn{
				ID:        int64(i + 1),
				Weight:    180.0 + float64(i),
				CreatedAt: fmt.Sprintf("2026-04-0%d 08:00:00", i+1),
				Source:    "manual",
				Notes:     fmt.Sprintf("day %d", i+1),
			})
		}

		var buf bytes.Buffer
		require.NoError(t, FormatList(&buf, "table", "kg", entries))

		out := buf.String()
		for _, e := range entries {
			assert.Contains(t, out, fmt.Sprintf("%.1f", e.Weight))
			assert.Contains(t, out, e.CreatedAt)
			assert.Contains(t, out, e.Notes)
		}
	})

	t.Run("unknown format defaults to table", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, FormatList(&buf, "yaml", "kg", []store.WeighIn{sampleWeighIn()}))

		out := buf.String()
		assert.Contains(t, out, "Timestamp")
		assert.Contains(t, out, "185.5")
	})
}
