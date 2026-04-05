package store

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	applog "github.com/julientant/weightlogr-cli/internal/logger"
)

const (
	OrderAsc  = "asc"
	OrderDesc = "desc"
)

// WeighIn represents a single weigh-in record.
type WeighIn struct {
	ID        int64   `json:"id"`
	Weight    float64 `json:"weight"`
	CreatedAt string  `json:"created_at"`
	Source    string  `json:"source"`
	Notes     string  `json:"notes"`
}

// ListOpts holds filtering/sorting options for listing weigh-ins.
type ListOpts struct {
	Since  string
	Until  string
	Source string
	Order  string // "asc" or "desc"
	Limit  int
}

// Store provides access to the weigh-ins table.
type Store struct {
	db *sql.DB
}

// New creates a Store backed by the given database connection.
func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// Insert adds a weigh-in and returns the created record.
func (s *Store) Insert(ctx context.Context, weight float64, createdAt, source, notes string) (WeighIn, error) {
	logger := applog.FromContext(ctx)

	query, args, err := sq.Insert("weigh_ins").
		Columns("weight", "created_at", "source", "notes").
		Values(weight, createdAt, source, notes).
		ToSql()
	if err != nil {
		return WeighIn{}, fmt.Errorf("build insert query: %w", err)
	}

	logger.DebugContext(ctx, "executing insert", "sql", query, "args", args)

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return WeighIn{}, fmt.Errorf("insert: %w", err)
	}

	rowID, err := result.LastInsertId()
	if err != nil {
		logger.WarnContext(ctx, "last insert id failed", "error", err)
	}
	logger.InfoContext(ctx, "weigh-in logged", "row_id", rowID, "weight", weight, "created_at", createdAt)

	return WeighIn{
		ID:        rowID,
		Weight:    weight,
		CreatedAt: createdAt,
		Source:    source,
		Notes:     notes,
	}, nil
}

// List retrieves weigh-ins matching the given options.
func (s *Store) List(ctx context.Context, opts ListOpts) ([]WeighIn, error) {
	logger := applog.FromContext(ctx)

	qb := sq.Select("id", "weight", "created_at", "source", "notes").
		From("weigh_ins")

	if opts.Since != "" {
		qb = qb.Where(sq.GtOrEq{"created_at": opts.Since})
	}
	if opts.Until != "" {
		qb = qb.Where(sq.Lt{"created_at": opts.Until})
	}
	if opts.Source != "" {
		qb = qb.Where(sq.Eq{"source": opts.Source})
	}

	if opts.Order == OrderAsc {
		qb = qb.OrderBy("created_at ASC")
	} else {
		qb = qb.OrderBy("created_at DESC")
	}

	if opts.Limit > 0 {
		qb = qb.Limit(uint64(opts.Limit))
	}

	query, args, err := qb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list query: %w", err)
	}

	logger.DebugContext(ctx, "executing query", "sql", query, "args", args)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			logger.ErrorContext(ctx, "rows close failed", "error", cerr)
		}
	}()

	var results []WeighIn
	for rows.Next() {
		var r WeighIn
		var source, notes *string
		if err := rows.Scan(&r.ID, &r.Weight, &r.CreatedAt, &source, &notes); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		if source != nil {
			r.Source = *source
		}
		if notes != nil {
			r.Notes = *notes
		}
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	logger.InfoContext(ctx, "weigh-ins retrieved", "count", len(results), "order", opts.Order)
	return results, nil
}
