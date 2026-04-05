package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	sq "github.com/Masterminds/squirrel"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/julientant/weightlogr-cli/internal/db"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List weigh-ins",
	RunE:  runList,
}

func init() {
	listCmd.Flags().String("since", "", "Start date/time inclusive (ISO 8601)")
	listCmd.Flags().String("until", "", "End date/time exclusive (ISO 8601)")
	listCmd.Flags().String("source", "", "Filter by source")
	listCmd.Flags().String("order", "desc", "Sort order: asc or desc")
	listCmd.Flags().Int("limit", 0, "Max rows to return (0 = unlimited)")

	rootCmd.AddCommand(listCmd)
}

type weighIn struct {
	ID        int64   `json:"id"`
	Weight    float64 `json:"weight"`
	CreatedAt string  `json:"created_at"`
	Source    string  `json:"source"`
	Notes     string  `json:"notes"`
}

func runList(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	logger := loggerFrom(ctx)

	logger.Debug("parsing list flags")

	tz, err := loadTimezone()
	if err != nil {
		return err
	}

	conn, err := db.Open(ctx, logger, viper.GetString("db"))
	if err != nil {
		return err
	}
	defer conn.Close()

	qb := sq.Select("id", "weight", "created_at", "source", "notes").
		From("weigh_ins")

	if since, _ := cmd.Flags().GetString("since"); since != "" {
		normalized := normalizeTimestamp(since, tz)
		logger.Debug("applying since filter", "raw", since, "normalized", normalized)
		qb = qb.Where(sq.GtOrEq{"created_at": normalized})
	}
	if until, _ := cmd.Flags().GetString("until"); until != "" {
		normalized := normalizeTimestamp(until, tz)
		logger.Debug("applying until filter", "raw", until, "normalized", normalized)
		qb = qb.Where(sq.Lt{"created_at": normalized})
	}
	if source, _ := cmd.Flags().GetString("source"); source != "" {
		logger.Debug("applying source filter", "source", source)
		qb = qb.Where(sq.Eq{"source": source})
	}

	order, _ := cmd.Flags().GetString("order")
	if order == "asc" {
		qb = qb.OrderBy("created_at ASC")
	} else {
		qb = qb.OrderBy("created_at DESC")
	}

	if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
		logger.Debug("applying limit", "limit", limit)
		qb = qb.Limit(uint64(limit))
	}

	query, args, err := qb.ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	logger.Debug("executing query", "sql", query, "args", args)

	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var results []weighIn
	for rows.Next() {
		var r weighIn
		var source, notes *string
		if err := rows.Scan(&r.ID, &r.Weight, &r.CreatedAt, &source, &notes); err != nil {
			return fmt.Errorf("scan: %w", err)
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
		return fmt.Errorf("rows: %w", err)
	}

	logger.Info("weigh-ins retrieved", "count", len(results), "order", order)

	return outputList(results)
}

func outputList(results []weighIn) error {
	format := viper.GetString("format")

	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(results)
	case "csv":
		w := csv.NewWriter(os.Stdout)
		w.Write([]string{"id", "weight", "created_at", "source", "notes"})
		for _, r := range results {
			w.Write([]string{
				fmt.Sprintf("%d", r.ID),
				fmt.Sprintf("%.1f", r.Weight),
				r.CreatedAt,
				r.Source,
				r.Notes,
			})
		}
		w.Flush()
		return w.Error()
	default:
		if len(results) == 0 {
			fmt.Println("No weigh-ins found.")
			return nil
		}
		fmt.Printf("%-20s | %6s | %-11s | %s\n", "Timestamp", "Weight", "Source", "Notes")
		fmt.Println("---------------------+--------+-------------+--------------------")
		for _, r := range results {
			notes := r.Notes
			if notes == "" {
				notes = "-"
			}
			fmt.Printf("%-20s | %6.1f | %-11s | %s\n", r.CreatedAt, r.Weight, r.Source, notes)
		}
		return nil
	}
}
