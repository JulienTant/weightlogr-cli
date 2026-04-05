package store

import (
	"context"
	"database/sql"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/julientant/weightlogr-cli/internal/db"
	applog "github.com/julientant/weightlogr-cli/internal/logger"
)

func openTestDB(t *testing.T) (*sql.DB, context.Context) {
	t.Helper()
	ctx := applog.WithContext(context.Background(), slog.Default())
	dbPath := filepath.Join(t.TempDir(), "test.db")
	conn, err := db.Open(ctx, dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	return conn, ctx
}

func seedThree(t *testing.T, s *Store, ctx context.Context) {
	t.Helper()
	_, err := s.Insert(ctx, 80.0, "2025-01-01T08:00:00Z", "scale", "a")
	require.NoError(t, err)
	_, err = s.Insert(ctx, 81.0, "2025-01-02T08:00:00Z", "manual", "b")
	require.NoError(t, err)
	_, err = s.Insert(ctx, 82.0, "2025-01-03T08:00:00Z", "scale", "c")
	require.NoError(t, err)
}

func TestInsert(t *testing.T) {
	t.Run("returns correct weigh-in", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)

		w, err := s.Insert(ctx, 82.5, "2025-01-15T08:00:00Z", "scale", "morning weigh-in")
		require.NoError(t, err)

		assert.Equal(t, int64(1), w.ID)
		assert.Equal(t, 82.5, w.Weight)
		assert.Equal(t, "2025-01-15T08:00:00Z", w.CreatedAt)
		assert.Equal(t, "scale", w.Source)
		assert.Equal(t, "morning weigh-in", w.Notes)
		assert.Equal(t, "2025-01-15T08:00:00Z", w.UpdatedAt)
	})

	t.Run("auto-incrementing IDs", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)

		w1, err := s.Insert(ctx, 80.0, "2025-01-01T08:00:00Z", "scale", "")
		require.NoError(t, err)
		w2, err := s.Insert(ctx, 81.0, "2025-01-02T08:00:00Z", "scale", "")
		require.NoError(t, err)
		w3, err := s.Insert(ctx, 82.0, "2025-01-03T08:00:00Z", "scale", "")
		require.NoError(t, err)

		assert.Equal(t, int64(1), w1.ID)
		assert.Equal(t, int64(2), w2.ID)
		assert.Equal(t, int64(3), w3.ID)
	})

	t.Run("empty notes and source", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)

		w, err := s.Insert(ctx, 79.0, "2025-02-01T08:00:00Z", "", "")
		require.NoError(t, err)

		assert.Equal(t, "", w.Source)
		assert.Equal(t, "", w.Notes)
	})

	t.Run("duplicate created_at returns error", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)

		_, err := s.Insert(ctx, 80.0, "2025-03-01T08:00:00Z", "scale", "first")
		require.NoError(t, err)

		_, err = s.Insert(ctx, 81.0, "2025-03-01T08:00:00Z", "manual", "duplicate")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "UNIQUE constraint failed")
	})

	t.Run("persists data via list", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)

		_, err := s.Insert(ctx, 85.0, "2025-04-01T08:00:00Z", "app", "persisted")
		require.NoError(t, err)

		results, err := s.List(ctx, ListOpts{})
		require.NoError(t, err)

		require.Len(t, results, 1)
		assert.Equal(t, 85.0, results[0].Weight)
		assert.Equal(t, "persisted", results[0].Notes)
	})
}

func TestList(t *testing.T) {
	t.Run("empty db", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)

		results, err := s.List(ctx, ListOpts{})
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("returns all entries", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)
		seedThree(t, s, ctx)

		results, err := s.List(ctx, ListOpts{})
		require.NoError(t, err)
		assert.Len(t, results, 3)
	})

	t.Run("since filter inclusive", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)
		seedThree(t, s, ctx)

		results, err := s.List(ctx, ListOpts{Since: "2025-01-02T08:00:00Z", Order: "asc"})
		require.NoError(t, err)

		require.Len(t, results, 2)
		assert.Equal(t, "2025-01-02T08:00:00Z", results[0].CreatedAt)
		assert.Equal(t, "2025-01-03T08:00:00Z", results[1].CreatedAt)
	})

	t.Run("until filter exclusive", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)
		seedThree(t, s, ctx)

		results, err := s.List(ctx, ListOpts{Until: "2025-01-03T08:00:00Z", Order: "asc"})
		require.NoError(t, err)

		require.Len(t, results, 2)
		assert.Equal(t, "2025-01-01T08:00:00Z", results[0].CreatedAt)
		assert.Equal(t, "2025-01-02T08:00:00Z", results[1].CreatedAt)
	})

	t.Run("since and until range", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)

		_, err := s.Insert(ctx, 80.0, "2025-01-01T08:00:00Z", "scale", "")
		require.NoError(t, err)
		_, err = s.Insert(ctx, 81.0, "2025-01-02T08:00:00Z", "scale", "")
		require.NoError(t, err)
		_, err = s.Insert(ctx, 82.0, "2025-01-03T08:00:00Z", "scale", "")
		require.NoError(t, err)
		_, err = s.Insert(ctx, 83.0, "2025-01-04T08:00:00Z", "scale", "")
		require.NoError(t, err)

		results, err := s.List(ctx, ListOpts{Since: "2025-01-02T08:00:00Z", Until: "2025-01-04T08:00:00Z", Order: "asc"})
		require.NoError(t, err)

		require.Len(t, results, 2)
		assert.Equal(t, "2025-01-02T08:00:00Z", results[0].CreatedAt)
		assert.Equal(t, "2025-01-03T08:00:00Z", results[1].CreatedAt)
	})

	t.Run("source filter", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)
		seedThree(t, s, ctx)

		results, err := s.List(ctx, ListOpts{Source: "scale"})
		require.NoError(t, err)

		require.Len(t, results, 2)
		for _, r := range results {
			assert.Equal(t, "scale", r.Source)
		}
	})

	t.Run("order asc vs desc", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)
		seedThree(t, s, ctx)

		asc, err := s.List(ctx, ListOpts{Order: "asc"})
		require.NoError(t, err)
		require.Len(t, asc, 3)
		assert.Equal(t, "2025-01-01T08:00:00Z", asc[0].CreatedAt)
		assert.Equal(t, "2025-01-03T08:00:00Z", asc[2].CreatedAt)

		desc, err := s.List(ctx, ListOpts{Order: "desc"})
		require.NoError(t, err)
		require.Len(t, desc, 3)
		assert.Equal(t, "2025-01-03T08:00:00Z", desc[0].CreatedAt)
		assert.Equal(t, "2025-01-01T08:00:00Z", desc[2].CreatedAt)

		defaultOrder, err := s.List(ctx, ListOpts{})
		require.NoError(t, err)
		assert.Equal(t, "2025-01-03T08:00:00Z", defaultOrder[0].CreatedAt)
	})

	t.Run("limit", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)
		seedThree(t, s, ctx)

		results, err := s.List(ctx, ListOpts{Limit: 2})
		require.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("limit zero returns all", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)
		seedThree(t, s, ctx)

		results, err := s.List(ctx, ListOpts{Limit: 0})
		require.NoError(t, err)
		assert.Len(t, results, 3)
	})

	t.Run("all opts combined", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)

		_, err := s.Insert(ctx, 80.0, "2025-01-01T08:00:00Z", "scale", "")
		require.NoError(t, err)
		_, err = s.Insert(ctx, 81.0, "2025-01-02T08:00:00Z", "manual", "")
		require.NoError(t, err)
		_, err = s.Insert(ctx, 82.0, "2025-01-03T08:00:00Z", "scale", "")
		require.NoError(t, err)
		_, err = s.Insert(ctx, 83.0, "2025-01-04T08:00:00Z", "scale", "")
		require.NoError(t, err)
		_, err = s.Insert(ctx, 84.0, "2025-01-05T08:00:00Z", "scale", "")
		require.NoError(t, err)

		results, err := s.List(ctx, ListOpts{
			Since:  "2025-01-02T08:00:00Z",
			Until:  "2025-01-05T08:00:00Z",
			Source: "scale",
			Order:  "asc",
			Limit:  2,
		})
		require.NoError(t, err)

		require.Len(t, results, 2)
		assert.Equal(t, "2025-01-03T08:00:00Z", results[0].CreatedAt)
		assert.Equal(t, "2025-01-04T08:00:00Z", results[1].CreatedAt)
	})

	t.Run("excludes soft-deleted rows", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)
		seedThree(t, s, ctx)

		err := s.Delete(ctx, 2)
		require.NoError(t, err)

		results, err := s.List(ctx, ListOpts{})
		require.NoError(t, err)
		assert.Len(t, results, 2)
		for _, r := range results {
			assert.NotEqual(t, int64(2), r.ID)
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)

		inserted, err := s.Insert(ctx, 80.0, "2025-01-01T08:00:00Z", "scale", "original")
		require.NoError(t, err)

		updated, err := s.Update(ctx, inserted.ID, 81.5, "manual", "corrected")
		require.NoError(t, err)

		assert.Equal(t, inserted.ID, updated.ID)
		assert.Equal(t, 81.5, updated.Weight)
		assert.Equal(t, "manual", updated.Source)
		assert.Equal(t, "corrected", updated.Notes)
		assert.Equal(t, inserted.CreatedAt, updated.CreatedAt)
		assert.NotEqual(t, inserted.UpdatedAt, updated.UpdatedAt)
	})

	t.Run("not found", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)

		_, err := s.Update(ctx, 999, 80.0, "scale", "")
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("already deleted", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)

		inserted, err := s.Insert(ctx, 80.0, "2025-01-01T08:00:00Z", "scale", "")
		require.NoError(t, err)

		err = s.Delete(ctx, inserted.ID)
		require.NoError(t, err)

		_, err = s.Update(ctx, inserted.ID, 81.0, "scale", "")
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("persists via list", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)

		inserted, err := s.Insert(ctx, 80.0, "2025-01-01T08:00:00Z", "scale", "original")
		require.NoError(t, err)

		_, err = s.Update(ctx, inserted.ID, 81.5, "manual", "updated")
		require.NoError(t, err)

		results, err := s.List(ctx, ListOpts{})
		require.NoError(t, err)

		require.Len(t, results, 1)
		assert.Equal(t, 81.5, results[0].Weight)
		assert.Equal(t, "updated", results[0].Notes)
	})
}

func TestDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)

		inserted, err := s.Insert(ctx, 80.0, "2025-01-01T08:00:00Z", "scale", "")
		require.NoError(t, err)

		err = s.Delete(ctx, inserted.ID)
		require.NoError(t, err)

		results, err := s.List(ctx, ListOpts{})
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("not found", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)

		err := s.Delete(ctx, 999)
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("already deleted", func(t *testing.T) {
		conn, ctx := openTestDB(t)
		s := New(conn)

		inserted, err := s.Insert(ctx, 80.0, "2025-01-01T08:00:00Z", "scale", "")
		require.NoError(t, err)

		err = s.Delete(ctx, inserted.ID)
		require.NoError(t, err)

		err = s.Delete(ctx, inserted.ID)
		require.ErrorIs(t, err, ErrNotFound)
	})
}
