package presentation

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/julientant/weightlogr-cli/internal/store"
	"github.com/julientant/weightlogr-cli/internal/version"
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
	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		r := sampleWeighIn()
		require.NoError(t, FormatInsert(&buf, FormatJSON, r))

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
		require.NoError(t, FormatInsert(&buf, FormatCSV, sampleWeighIn()))

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
		require.NoError(t, FormatInsert(&buf, FormatCSV, r))

		records, err := csv.NewReader(strings.NewReader(buf.String())).ReadAll()
		require.NoError(t, err)

		require.Len(t, records, 2)
		assert.Equal(t, "felt tired, sore", records[1][4])
	})

	t.Run("unsupported format returns error", func(t *testing.T) {
		var buf bytes.Buffer
		err := FormatInsert(&buf, "xml", sampleWeighIn())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported format")
	})
}

func TestFormatList(t *testing.T) {
	twoEntries := []store.WeighIn{
		{ID: 1, Weight: 185.0, CreatedAt: "2026-04-01 08:00:00", Source: "manual", Notes: "good day"},
		{ID: 2, Weight: 184.5, CreatedAt: "2026-04-02 08:00:00", Source: "scale", Notes: ""},
	}

	t.Run("json empty", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, FormatList(&buf, FormatJSON, nil))
		assert.Equal(t, "null\n", buf.String())
	})

	t.Run("json with entries", func(t *testing.T) {
		entries := []store.WeighIn{
			{ID: 1, Weight: 185.0, CreatedAt: "2026-04-01 08:00:00", Source: "manual", Notes: "a"},
			{ID: 2, Weight: 184.5, CreatedAt: "2026-04-02 08:00:00", Source: "scale", Notes: "b"},
		}

		var buf bytes.Buffer
		require.NoError(t, FormatList(&buf, FormatJSON, entries))

		var decoded []store.WeighIn
		require.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))

		require.Len(t, decoded, 2)
		assert.Equal(t, int64(1), decoded[0].ID)
		assert.InDelta(t, 185.0, decoded[0].Weight, 0.001)
		assert.Equal(t, "a", decoded[0].Notes)
	})

	t.Run("csv", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, FormatList(&buf, FormatCSV, twoEntries))

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
		require.NoError(t, FormatList(&buf, FormatCSV, entries))

		records, err := csv.NewReader(strings.NewReader(buf.String())).ReadAll()
		require.NoError(t, err)

		require.Len(t, records, 3)
		assert.Equal(t, "tired, sore", records[1][4])
		assert.Equal(t, `she said "wow"`, records[2][4])
	})

	t.Run("unsupported format returns error", func(t *testing.T) {
		var buf bytes.Buffer
		err := FormatList(&buf, "yaml", []store.WeighIn{sampleWeighIn()})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported format")
	})
}

func TestFormatVersion(t *testing.T) {
	info := version.BuildInfo{
		Version: "1.2.3",
		Commit:  "abc123",
		Date:    "2026-04-05",
	}

	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, FormatVersion(&buf, FormatJSON, info))

		var decoded version.BuildInfo
		require.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))

		assert.Equal(t, info, decoded)
	})

	t.Run("csv", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, FormatVersion(&buf, FormatCSV, info))

		records, err := csv.NewReader(strings.NewReader(buf.String())).ReadAll()
		require.NoError(t, err)

		require.Len(t, records, 2)
		assert.Equal(t, []string{"version", "commit", "date"}, records[0])
		assert.Equal(t, []string{"1.2.3", "abc123", "2026-04-05"}, records[1])
	})

	t.Run("unsupported format returns error", func(t *testing.T) {
		var buf bytes.Buffer
		err := FormatVersion(&buf, "xml", info)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported format")
	})
}
