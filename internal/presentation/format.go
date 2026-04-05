package presentation

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"

	"github.com/julientant/weightlogr-cli/internal/store"
	"github.com/julientant/weightlogr-cli/internal/version"
)

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
		if err := cw.Write([]string{"id", "weight", "created_at", "source", "notes"}); err != nil {
			return fmt.Errorf("csv header: %w", err)
		}
		if err := cw.Write([]string{
			fmt.Sprintf("%d", r.ID),
			fmt.Sprintf("%.1f", r.Weight),
			r.CreatedAt,
			r.Source,
			r.Notes,
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
func FormatList(w io.Writer, format string, results []store.WeighIn) error {
	switch format {
	case FormatJSON:
		if err := json.NewEncoder(w).Encode(results); err != nil {
			return fmt.Errorf("json encode: %w", err)
		}
		return nil
	case FormatCSV:
		cw := csv.NewWriter(w)
		if err := cw.Write([]string{"id", "weight", "created_at", "source", "notes"}); err != nil {
			return fmt.Errorf("csv header: %w", err)
		}
		for _, r := range results {
			if err := cw.Write([]string{
				fmt.Sprintf("%d", r.ID),
				fmt.Sprintf("%.1f", r.Weight),
				r.CreatedAt,
				r.Source,
				r.Notes,
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
