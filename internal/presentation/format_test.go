package presentation

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/julientant/weightlogr-cli/internal/store"
	"github.com/julientant/weightlogr-cli/internal/version"
)

func sampleWeighIn() store.WeighIn {
	return store.WeighIn{
		ID:        1,
		Weight:    185.5,
		CreatedAt: "2026-04-05T08:30:00Z",
		Source:    "manual",
		Notes:     "morning",
		UpdatedAt: "2026-04-05T08:30:00Z",
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
		assert.Equal(t, []string{"id", "weight", "created_at", "source", "notes", "updated_at"}, records[0])
		assert.Equal(t, "1", records[1][0])
		assert.Equal(t, "185.5", records[1][1])
		assert.Equal(t, "2026-04-05T08:30:00Z", records[1][2])
		assert.Equal(t, "manual", records[1][3])
		assert.Equal(t, "morning", records[1][4])
		assert.Equal(t, "2026-04-05T08:30:00Z", records[1][5])
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
		{ID: 1, Weight: 185.0, CreatedAt: "2026-04-01T08:00:00Z", Source: "manual", Notes: "good day", UpdatedAt: "2026-04-01T08:00:00Z"},
		{ID: 2, Weight: 184.5, CreatedAt: "2026-04-02T08:00:00Z", Source: "scale", Notes: "", UpdatedAt: "2026-04-02T08:00:00Z"},
	}

	t.Run("json empty", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, FormatList(&buf, FormatJSON, time.UTC, nil))
		assert.Equal(t, "null\n", buf.String())
	})

	t.Run("json with entries", func(t *testing.T) {
		entries := []store.WeighIn{
			{ID: 1, Weight: 185.0, CreatedAt: "2026-04-01T08:00:00Z", Source: "manual", Notes: "a", UpdatedAt: "2026-04-01T08:00:00Z"},
			{ID: 2, Weight: 184.5, CreatedAt: "2026-04-02T08:00:00Z", Source: "scale", Notes: "b", UpdatedAt: "2026-04-02T08:00:00Z"},
		}

		var buf bytes.Buffer
		require.NoError(t, FormatList(&buf, FormatJSON, time.UTC, entries))

		var decoded []store.WeighIn
		require.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))

		require.Len(t, decoded, 2)
		assert.Equal(t, int64(1), decoded[0].ID)
		assert.InDelta(t, 185.0, decoded[0].Weight, 0.001)
		assert.Equal(t, "a", decoded[0].Notes)
	})

	t.Run("json with timezone conversion", func(t *testing.T) {
		entries := []store.WeighIn{
			{ID: 1, Weight: 185.0, CreatedAt: "2026-04-01T15:00:00Z", Source: "manual", Notes: "a", UpdatedAt: "2026-04-01T15:00:00Z"},
		}

		phoenix, err := time.LoadLocation("America/Phoenix")
		require.NoError(t, err)

		var buf bytes.Buffer
		require.NoError(t, FormatList(&buf, FormatJSON, phoenix, entries))

		var decoded []store.WeighIn
		require.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))

		require.Len(t, decoded, 1)
		assert.Equal(t, "2026-04-01T08:00:00-07:00", decoded[0].CreatedAt)
	})

	t.Run("csv", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, FormatList(&buf, FormatCSV, time.UTC, twoEntries))

		records, err := csv.NewReader(strings.NewReader(buf.String())).ReadAll()
		require.NoError(t, err)

		require.Len(t, records, 3)
		assert.Equal(t, []string{"id", "weight", "created_at", "source", "notes", "updated_at"}, records[0])
		assert.Equal(t, "185.0", records[1][1])
		assert.Equal(t, "184.5", records[2][1])
	})

	t.Run("csv with timezone conversion", func(t *testing.T) {
		entries := []store.WeighIn{
			{ID: 1, Weight: 185.0, CreatedAt: "2026-04-01T15:00:00Z", Source: "manual", Notes: "afternoon", UpdatedAt: "2026-04-01T15:00:00Z"},
		}

		phoenix, err := time.LoadLocation("America/Phoenix")
		require.NoError(t, err)

		var buf bytes.Buffer
		require.NoError(t, FormatList(&buf, FormatCSV, phoenix, entries))

		records, err := csv.NewReader(strings.NewReader(buf.String())).ReadAll()
		require.NoError(t, err)

		require.Len(t, records, 2)
		assert.Equal(t, "2026-04-01T08:00:00-07:00", records[1][2])
	})

	t.Run("csv special characters in notes", func(t *testing.T) {
		entries := []store.WeighIn{
			{ID: 1, Weight: 185.0, CreatedAt: "2026-04-01T08:00:00Z", Source: "manual", Notes: "tired, sore", UpdatedAt: "2026-04-01T08:00:00Z"},
			{ID: 2, Weight: 184.5, CreatedAt: "2026-04-02T08:00:00Z", Source: "scale", Notes: `she said "wow"`, UpdatedAt: "2026-04-02T08:00:00Z"},
		}

		var buf bytes.Buffer
		require.NoError(t, FormatList(&buf, FormatCSV, time.UTC, entries))

		records, err := csv.NewReader(strings.NewReader(buf.String())).ReadAll()
		require.NoError(t, err)

		require.Len(t, records, 3)
		assert.Equal(t, "tired, sore", records[1][4])
		assert.Equal(t, `she said "wow"`, records[2][4])
	})

	t.Run("unsupported format returns error", func(t *testing.T) {
		var buf bytes.Buffer
		err := FormatList(&buf, "yaml", time.UTC, []store.WeighIn{sampleWeighIn()})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported format")
	})
}

func TestFormatUpdate(t *testing.T) {
	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		r := sampleWeighIn()
		r.UpdatedAt = "2026-04-06T10:00:00Z"
		require.NoError(t, FormatUpdate(&buf, FormatJSON, r))

		var decoded store.WeighIn
		require.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))

		assert.Equal(t, r.UpdatedAt, decoded.UpdatedAt)
		assert.Equal(t, r.Weight, decoded.Weight)
	})

	t.Run("csv", func(t *testing.T) {
		var buf bytes.Buffer
		r := sampleWeighIn()
		r.UpdatedAt = "2026-04-06T10:00:00Z"
		require.NoError(t, FormatUpdate(&buf, FormatCSV, r))

		records, err := csv.NewReader(strings.NewReader(buf.String())).ReadAll()
		require.NoError(t, err)

		require.Len(t, records, 2)
		assert.Equal(t, "updated_at", records[0][5])
		assert.Equal(t, "2026-04-06T10:00:00Z", records[1][5])
	})
}

func TestFormatDelete(t *testing.T) {
	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, FormatDelete(&buf, FormatJSON, 42))

		var decoded DeleteResult
		require.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))

		assert.Equal(t, int64(42), decoded.ID)
		assert.True(t, decoded.Deleted)
	})

	t.Run("csv", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, FormatDelete(&buf, FormatCSV, 42))

		records, err := csv.NewReader(strings.NewReader(buf.String())).ReadAll()
		require.NoError(t, err)

		require.Len(t, records, 2)
		assert.Equal(t, []string{"id", "deleted"}, records[0])
		assert.Equal(t, []string{"42", "true"}, records[1])
	})

	t.Run("unsupported format returns error", func(t *testing.T) {
		var buf bytes.Buffer
		err := FormatDelete(&buf, "xml", 1)
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
