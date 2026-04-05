package presentation

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/julientant/weightlogr-cli/internal/store"
	"github.com/julientant/weightlogr-cli/internal/version"
)

func convertTimestamp(rfc3339 string, loc *time.Location) (string, error) {
	t, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		return rfc3339, fmt.Errorf("parse timestamp %q: %w", rfc3339, err)
	}
	return t.In(loc).Format(time.RFC3339), nil
}

const (
	FormatJSON = "json"
	FormatCSV  = "csv"
)

// FormatInsert writes a single weigh-in result in the given format.
func FormatInsert(w io.Writer, format string, r store.WeighIn) error {
	switch format {
	case FormatJSON:
		if err := json.NewEncoder(w).Encode(r); err != nil {
			return fmt.Errorf("json encode: %w", err)
		}
		return nil
	case FormatCSV:
		cw := csv.NewWriter(w)
		if err := cw.Write([]string{"id", "weight", "created_at", "source", "notes", "updated_at"}); err != nil {
			return fmt.Errorf("csv header: %w", err)
		}
		if err := cw.Write([]string{
			fmt.Sprintf("%d", r.ID),
			fmt.Sprintf("%.1f", r.Weight),
			r.CreatedAt,
			r.Source,
			r.Notes,
			r.UpdatedAt,
		}); err != nil {
			return fmt.Errorf("csv row: %w", err)
		}
		cw.Flush()
		if err := cw.Error(); err != nil {
			return fmt.Errorf("csv flush: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// FormatList writes a list of weigh-ins in the given format.
// Timestamps are converted to loc before formatting.
func FormatList(w io.Writer, format string, loc *time.Location, results []store.WeighIn) error {
	switch format {
	case FormatJSON:
		if results == nil {
			if err := json.NewEncoder(w).Encode(results); err != nil {
				return fmt.Errorf("json encode: %w", err)
			}
			return nil
		}
		converted := make([]store.WeighIn, len(results))
		for i, r := range results {
			ts, err := convertTimestamp(r.CreatedAt, loc)
			if err != nil {
				return fmt.Errorf("convert timestamp: %w", err)
			}
			r.CreatedAt = ts
			updatedTs, err := convertTimestamp(r.UpdatedAt, loc)
			if err != nil {
				return fmt.Errorf("convert updated_at timestamp: %w", err)
			}
			r.UpdatedAt = updatedTs
			converted[i] = r
		}
		if err := json.NewEncoder(w).Encode(converted); err != nil {
			return fmt.Errorf("json encode: %w", err)
		}
		return nil
	case FormatCSV:
		cw := csv.NewWriter(w)
		if err := cw.Write([]string{"id", "weight", "created_at", "source", "notes", "updated_at"}); err != nil {
			return fmt.Errorf("csv header: %w", err)
		}
		for _, r := range results {
			ts, err := convertTimestamp(r.CreatedAt, loc)
			if err != nil {
				return fmt.Errorf("convert timestamp: %w", err)
			}
			updatedTs, err := convertTimestamp(r.UpdatedAt, loc)
			if err != nil {
				return fmt.Errorf("convert updated_at timestamp: %w", err)
			}
			if err := cw.Write([]string{
				fmt.Sprintf("%d", r.ID),
				fmt.Sprintf("%.1f", r.Weight),
				ts,
				r.Source,
				r.Notes,
				updatedTs,
			}); err != nil {
				return fmt.Errorf("csv row: %w", err)
			}
		}
		cw.Flush()
		if err := cw.Error(); err != nil {
			return fmt.Errorf("csv flush: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// FormatUpdate writes a single updated weigh-in result in the given format.
func FormatUpdate(w io.Writer, format string, r store.WeighIn) error {
	return FormatInsert(w, format, r)
}

// DeleteResult represents the output of a delete operation.
type DeleteResult struct {
	ID      int64 `json:"id"`
	Deleted bool  `json:"deleted"`
}

// FormatDelete writes a delete confirmation in the given format.
func FormatDelete(w io.Writer, format string, id int64) error {
	r := DeleteResult{ID: id, Deleted: true}
	switch format {
	case FormatJSON:
		if err := json.NewEncoder(w).Encode(r); err != nil {
			return fmt.Errorf("json encode: %w", err)
		}
		return nil
	case FormatCSV:
		cw := csv.NewWriter(w)
		if err := cw.Write([]string{"id", "deleted"}); err != nil {
			return fmt.Errorf("csv header: %w", err)
		}
		if err := cw.Write([]string{fmt.Sprintf("%d", r.ID), "true"}); err != nil {
			return fmt.Errorf("csv row: %w", err)
		}
		cw.Flush()
		if err := cw.Error(); err != nil {
			return fmt.Errorf("csv flush: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// FormatVersion writes build info in the given format.
func FormatVersion(w io.Writer, format string, info version.BuildInfo) error {
	switch format {
	case FormatJSON:
		if err := json.NewEncoder(w).Encode(info); err != nil {
			return fmt.Errorf("json encode: %w", err)
		}
		return nil
	case FormatCSV:
		cw := csv.NewWriter(w)
		if err := cw.Write([]string{"version", "commit", "date"}); err != nil {
			return fmt.Errorf("csv header: %w", err)
		}
		if err := cw.Write([]string{info.Version, info.Commit, info.Date}); err != nil {
			return fmt.Errorf("csv row: %w", err)
		}
		cw.Flush()
		if err := cw.Error(); err != nil {
			return fmt.Errorf("csv flush: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}
