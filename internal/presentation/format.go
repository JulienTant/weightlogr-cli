package presentation

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"

	"github.com/julientant/weightlogr-cli/internal/store"
)

const (
	FormatJSON  = "json"
	FormatCSV   = "csv"
	FormatTable = "table"

	tableHeader    = "%-20s | %6s | %-11s | %s\n"
	tableSeparator = "---------------------+--------+-------------+--------------------"
	tableRow       = "%-20s | %6.1f | %-11s | %s\n"
	emptyNotes     = "-"
)

// FormatInsert writes a single weigh-in result in the given format.
func FormatInsert(w io.Writer, format, unit string, r store.WeighIn) error {
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
		if _, err := fmt.Fprintf(w, "Logged %.1f %s at %s (row id %d).\n", r.Weight, unit, r.CreatedAt, r.ID); err != nil {
			return fmt.Errorf("write table output: %w", err)
		}
		return nil
	}
}

// FormatList writes a list of weigh-ins in the given format.
func FormatList(w io.Writer, format, unit string, results []store.WeighIn) error {
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
		if len(results) == 0 {
			if _, err := fmt.Fprintln(w, "No weigh-ins found."); err != nil {
				return fmt.Errorf("write empty message: %w", err)
			}
			return nil
		}
		if _, err := fmt.Fprintf(w, tableHeader, "Timestamp", "Weight", "Source", "Notes"); err != nil {
			return fmt.Errorf("write table header: %w", err)
		}
		if _, err := fmt.Fprintln(w, tableSeparator); err != nil {
			return fmt.Errorf("write table separator: %w", err)
		}
		for _, r := range results {
			notes := r.Notes
			if notes == "" {
				notes = emptyNotes
			}
			if _, err := fmt.Fprintf(w, tableRow, r.CreatedAt, r.Weight, r.Source, notes); err != nil {
				return fmt.Errorf("write table row: %w", err)
			}
		}
		return nil
	}
}
