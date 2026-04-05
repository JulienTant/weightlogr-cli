package e2e_test

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/julientant/weightlogr-cli/internal/version"
	"github.com/julientant/weightlogr-cli/pkg/models"
)

var (
	builtBin  string
	buildOnce sync.Once
	buildErr  error
)

type testEnv struct {
	t       *testing.T
	bin     string
	dbPath  string
	logPath string
}

func setup(t *testing.T) *testEnv {
	t.Helper()

	bin := buildBinary(t)

	dbPath := filepath.Join(t.TempDir(), "test.db")
	logPath := filepath.Join(t.TempDir(), "test.log")

	return &testEnv{t: t, bin: bin, dbPath: dbPath, logPath: logPath}
}

func buildBinary(t *testing.T) string {
	t.Helper()

	buildOnce.Do(func() {
		// e2e-tests/ is one level below the project root
		root, err := filepath.Abs("..")
		if err != nil {
			buildErr = err
			return
		}

		dir, err := os.MkdirTemp("", "weightlogr-e2e-*")
		if err != nil {
			buildErr = err
			return
		}

		builtBin = filepath.Join(dir, "weightlogr-cli")
		cmd := exec.Command("go", "build", "-o", builtBin, ".")
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		if err != nil {
			buildErr = fmt.Errorf("build failed: %s: %w", string(out), err)
		}
	})

	require.NoError(t, buildErr)
	return builtBin
}

func (e *testEnv) run(args ...string) (string, error) {
	e.t.Helper()
	allArgs := append(args, "--db", e.dbPath, "--log-file", e.logPath, "--log-level", "debug")
	cmd := exec.Command(e.bin, allArgs...)
	out, err := cmd.CombinedOutput()
	return strings.TrimRight(string(out), "\n"), err
}

func (e *testEnv) mustRun(args ...string) string {
	e.t.Helper()
	out, err := e.run(args...)
	require.NoError(e.t, err, "command failed: %s\nargs: %v", out, args)
	return out
}

func (e *testEnv) seed() {
	e.t.Helper()
	e.mustRun("insert", "185.2", "--timestamp", "2026-04-05T15:00:00Z", "--notes", "morning")
	e.mustRun("insert", "183.8", "--timestamp", "2026-04-04T14:30:00Z", "--source", "gym-check", "--notes", "after gym")
	e.mustRun("insert", "184.0", "--timestamp", "2026-04-06T16:00:00Z", "--notes", "before lunch, felt light")
}

func parseJSONList(t *testing.T, raw string) []models.WeighIn {
	t.Helper()
	var result []models.WeighIn
	require.NoError(t, json.Unmarshal([]byte(raw), &result))
	return result
}

func parseJSONOne(t *testing.T, raw string) models.WeighIn {
	t.Helper()
	var result models.WeighIn
	require.NoError(t, json.Unmarshal([]byte(raw), &result))
	return result
}

func parseCSV(t *testing.T, raw string) [][]string {
	t.Helper()
	records, err := csv.NewReader(strings.NewReader(raw)).ReadAll()
	require.NoError(t, err)
	return records
}

func TestListJSON(t *testing.T) {
	env := setup(t)
	env.seed()

	out := env.mustRun("list", "--format", "json")
	results := parseJSONList(t, out)

	require.Len(t, results, 3)
	// Default order is desc by created_at
	assert.Equal(t, int64(3), results[0].ID)
	assert.InDelta(t, 184.0, results[0].Weight, 0.001)
	assert.Equal(t, "2026-04-06T16:00:00Z", results[0].CreatedAt)
	assert.Equal(t, "daily-check", results[0].Source)
	assert.Equal(t, "before lunch, felt light", results[0].Notes)
	assert.Equal(t, results[0].CreatedAt, results[0].UpdatedAt)

	assert.Equal(t, int64(1), results[1].ID)
	assert.InDelta(t, 185.2, results[1].Weight, 0.001)

	assert.Equal(t, int64(2), results[2].ID)
	assert.InDelta(t, 183.8, results[2].Weight, 0.001)
	assert.Equal(t, "gym-check", results[2].Source)
}

func TestListCSV(t *testing.T) {
	env := setup(t)
	env.seed()

	out := env.mustRun("list", "--format", "csv")
	records := parseCSV(t, out)

	require.Len(t, records, 4) // header + 3 rows
	assert.Equal(t, []string{"id", "weight", "created_at", "source", "notes", "updated_at"}, records[0])
	assert.Equal(t, "3", records[1][0])
	assert.Equal(t, "184.0", records[1][1])
	assert.Equal(t, "2026-04-06T16:00:00Z", records[1][2])
}

func TestListFilters(t *testing.T) {
	env := setup(t)
	env.seed()

	t.Run("since", func(t *testing.T) {
		out := env.mustRun("list", "--format", "json", "--since", "2026-04-05T00:00:00Z")
		results := parseJSONList(t, out)
		require.Len(t, results, 2)
		assert.Equal(t, int64(3), results[0].ID)
		assert.Equal(t, int64(1), results[1].ID)
	})

	t.Run("until", func(t *testing.T) {
		out := env.mustRun("list", "--format", "json", "--until", "2026-04-05T00:00:00Z")
		results := parseJSONList(t, out)
		require.Len(t, results, 1)
		assert.Equal(t, int64(2), results[0].ID)
	})

	t.Run("source", func(t *testing.T) {
		out := env.mustRun("list", "--format", "json", "--source", "gym-check")
		results := parseJSONList(t, out)
		require.Len(t, results, 1)
		assert.Equal(t, int64(2), results[0].ID)
		assert.Equal(t, "gym-check", results[0].Source)
	})
}

func TestListOrderAndLimit(t *testing.T) {
	env := setup(t)
	env.seed()

	t.Run("order asc", func(t *testing.T) {
		out := env.mustRun("list", "--format", "json", "--order", "asc")
		results := parseJSONList(t, out)
		require.Len(t, results, 3)
		assert.Equal(t, int64(2), results[0].ID)
		assert.Equal(t, int64(1), results[1].ID)
		assert.Equal(t, int64(3), results[2].ID)
	})

	t.Run("limit", func(t *testing.T) {
		out := env.mustRun("list", "--format", "json", "--limit", "1")
		results := parseJSONList(t, out)
		require.Len(t, results, 1)
		assert.Equal(t, int64(3), results[0].ID)
	})
}

func TestListTimezone(t *testing.T) {
	env := setup(t)
	env.seed()

	t.Run("json", func(t *testing.T) {
		out := env.mustRun("list", "--format", "json", "--timezone", "America/Phoenix")
		results := parseJSONList(t, out)
		require.Len(t, results, 3)
		assert.Equal(t, "2026-04-06T09:00:00-07:00", results[0].CreatedAt)
		assert.Equal(t, "2026-04-06T09:00:00-07:00", results[0].UpdatedAt)
		assert.Equal(t, "2026-04-05T08:00:00-07:00", results[1].CreatedAt)
		assert.Equal(t, "2026-04-04T07:30:00-07:00", results[2].CreatedAt)
	})

	t.Run("csv", func(t *testing.T) {
		out := env.mustRun("list", "--format", "csv", "--timezone", "America/Phoenix")
		records := parseCSV(t, out)
		require.Len(t, records, 4)
		assert.Equal(t, "2026-04-06T09:00:00-07:00", records[1][2])
		assert.Equal(t, "2026-04-06T09:00:00-07:00", records[1][5]) // updated_at
	})
}

func TestInsertCSV(t *testing.T) {
	env := setup(t)

	out := env.mustRun("insert", "187.0", "--timestamp", "2026-04-08T10:00:00Z", "--format", "csv")
	records := parseCSV(t, out)

	require.Len(t, records, 2)
	assert.Equal(t, []string{"id", "weight", "created_at", "source", "notes", "updated_at"}, records[0])
	assert.Equal(t, "1", records[1][0])
	assert.Equal(t, "187.0", records[1][1])
	assert.Equal(t, "2026-04-08T10:00:00Z", records[1][2])
	assert.Equal(t, "daily-check", records[1][3])
	assert.Equal(t, "", records[1][4])
	assert.Equal(t, "2026-04-08T10:00:00Z", records[1][5])
}

func TestInsertUTCStorage(t *testing.T) {
	env := setup(t)

	out := env.mustRun("insert", "186.0", "--timestamp", "2026-04-07T08:00:00-07:00")
	result := parseJSONOne(t, out)

	assert.Equal(t, "2026-04-07T15:00:00Z", result.CreatedAt)
	assert.Equal(t, "2026-04-07T15:00:00Z", result.UpdatedAt)
}

func TestVersion(t *testing.T) {
	env := setup(t)

	t.Run("json", func(t *testing.T) {
		out := env.mustRun("version", "--format", "json")
		var info version.BuildInfo
		require.NoError(t, json.Unmarshal([]byte(out), &info))
		assert.Equal(t, "dev", info.Version)
		assert.Equal(t, "none", info.Commit)
		assert.Equal(t, "unknown", info.Date)
	})

	t.Run("csv", func(t *testing.T) {
		out := env.mustRun("version", "--format", "csv")
		records := parseCSV(t, out)
		require.Len(t, records, 2)
		assert.Equal(t, []string{"version", "commit", "date"}, records[0])
		assert.Equal(t, []string{"dev", "none", "unknown"}, records[1])
	})
}

func TestEmptyDB(t *testing.T) {
	env := setup(t)

	t.Run("json", func(t *testing.T) {
		out := env.mustRun("list", "--format", "json")
		assert.Equal(t, "null", out)
	})

	t.Run("csv", func(t *testing.T) {
		out := env.mustRun("list", "--format", "csv")
		records := parseCSV(t, out)
		require.Len(t, records, 1) // header only
		assert.Equal(t, []string{"id", "weight", "created_at", "source", "notes", "updated_at"}, records[0])
	})
}

func TestUpdate(t *testing.T) {
	env := setup(t)
	env.seed()

	t.Run("updates weight and fields", func(t *testing.T) {
		out := env.mustRun("update", "1", "190.0", "--source", "corrected", "--notes", "re-weighed")
		result := parseJSONOne(t, out)

		assert.Equal(t, int64(1), result.ID)
		assert.InDelta(t, 190.0, result.Weight, 0.001)
		assert.Equal(t, "corrected", result.Source)
		assert.Equal(t, "re-weighed", result.Notes)
		assert.Equal(t, "2026-04-05T15:00:00Z", result.CreatedAt)
		assert.NotEqual(t, result.CreatedAt, result.UpdatedAt)
	})

	t.Run("persists via list", func(t *testing.T) {
		out := env.mustRun("list", "--format", "json", "--order", "asc")
		results := parseJSONList(t, out)

		var found bool
		for _, r := range results {
			if r.ID == 1 {
				assert.InDelta(t, 190.0, r.Weight, 0.001)
				assert.Equal(t, "corrected", r.Source)
				found = true
			}
		}
		assert.True(t, found, "updated entry not found in list")
	})

	t.Run("non-existent id fails", func(t *testing.T) {
		_, err := env.run("update", "999", "80.0")
		require.Error(t, err)
	})

	t.Run("invalid id fails", func(t *testing.T) {
		_, err := env.run("update", "not-a-number", "80.0")
		require.Error(t, err)
	})
}

func TestDelete(t *testing.T) {
	env := setup(t)
	env.seed()

	t.Run("returns confirmation", func(t *testing.T) {
		out := env.mustRun("delete", "2")
		var result models.DeleteResult
		require.NoError(t, json.Unmarshal([]byte(out), &result))
		assert.Equal(t, int64(2), result.ID)
		assert.True(t, result.Deleted)
	})

	t.Run("excluded from list", func(t *testing.T) {
		out := env.mustRun("list", "--format", "json")
		results := parseJSONList(t, out)
		for _, r := range results {
			assert.NotEqual(t, int64(2), r.ID, "deleted entry should not appear in list")
		}
	})

	t.Run("non-existent id fails", func(t *testing.T) {
		_, err := env.run("delete", "999")
		require.Error(t, err)
	})

	t.Run("invalid id fails", func(t *testing.T) {
		_, err := env.run("delete", "not-a-number")
		require.Error(t, err)
	})

	t.Run("already deleted fails", func(t *testing.T) {
		_, err := env.run("delete", "2")
		require.Error(t, err)
	})
}

func TestErrorHandling(t *testing.T) {
	env := setup(t)
	env.seed()

	t.Run("invalid weight", func(t *testing.T) {
		_, err := env.run("insert", "not-a-number")
		require.Error(t, err)
	})

	t.Run("duplicate timestamp", func(t *testing.T) {
		_, err := env.run("insert", "180.0", "--timestamp", "2026-04-05T15:00:00Z")
		require.Error(t, err)
	})

	t.Run("missing weight arg", func(t *testing.T) {
		_, err := env.run("insert")
		require.Error(t, err)
	})

	t.Run("non-RFC3339 insert timestamp", func(t *testing.T) {
		_, err := env.run("insert", "180.0", "--timestamp", "2026-04-05 08:00")
		require.Error(t, err)
	})

	t.Run("non-RFC3339 since", func(t *testing.T) {
		_, err := env.run("list", "--since", "2026-04-05 08:00")
		require.Error(t, err)
	})

	t.Run("invalid timezone", func(t *testing.T) {
		_, err := env.run("list", "--timezone", "Invalid/Zone")
		require.Error(t, err)
	})
}
