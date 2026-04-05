package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/julientant/weightlogr-cli/internal/db"
)

var insertCmd = &cobra.Command{
	Use:   "insert <weight>",
	Short: "Log a weigh-in",
	Args:  cobra.ExactArgs(1),
	RunE:  runInsert,
}

func init() {
	insertCmd.Flags().String("timestamp", "", "ISO 8601 timestamp (defaults to now)")
	insertCmd.Flags().String("source", "daily-check", "Source label")
	insertCmd.Flags().String("notes", "", "Optional notes")

	rootCmd.AddCommand(insertCmd)
}

func runInsert(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	logger := loggerFrom(ctx)

	logger.Debug("parsing insert arguments", "raw_weight", args[0])

	tz, err := loadTimezone()
	if err != nil {
		return err
	}
	logger.Debug("timezone loaded", "tz", tz.String())

	var weight float64
	if _, err := fmt.Sscanf(args[0], "%f", &weight); err != nil {
		return fmt.Errorf("invalid weight: %w", err)
	}

	ts, _ := cmd.Flags().GetString("timestamp")
	createdAt := normalizeTimestamp(ts, tz)
	logger.Debug("timestamp resolved", "raw", ts, "normalized", createdAt)

	source, _ := cmd.Flags().GetString("source")
	notes, _ := cmd.Flags().GetString("notes")

	logger.Info("inserting weigh-in",
		"weight", weight,
		"created_at", createdAt,
		"source", source,
	)

	conn, err := db.Open(ctx, logger, viper.GetString("db"))
	if err != nil {
		return err
	}
	defer conn.Close()

	query, queryArgs, err := sq.Insert("weigh_ins").
		Columns("weight", "created_at", "source", "notes").
		Values(weight, createdAt, source, notes).
		ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	logger.Debug("executing insert", "sql", query, "args", queryArgs)

	result, err := conn.ExecContext(ctx, query, queryArgs...)
	if err != nil {
		return fmt.Errorf("insert: %w", err)
	}

	rowID, _ := result.LastInsertId()
	logger.Info("weigh-in logged", "row_id", rowID, "weight", weight, "created_at", createdAt)

	return outputInsertResult(weight, createdAt, source, notes, rowID)
}

func normalizeTimestamp(value string, tz *time.Location) string {
	if value == "" {
		return time.Now().In(tz).Format("2006-01-02 15:04:05")
	}

	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t, err = time.ParseInLocation("2006-01-02T15:04", value, tz)
		if err != nil {
			t, _ = time.ParseInLocation("2006-01-02T15:04:05", value, tz)
		}
	}

	return t.In(tz).Format("2006-01-02 15:04:05")
}

func outputInsertResult(weight float64, createdAt, source, notes string, rowID int64) error {
	format := viper.GetString("format")

	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"id":         rowID,
			"weight":     weight,
			"created_at": createdAt,
			"source":     source,
			"notes":      notes,
		})
	case "csv":
		w := csv.NewWriter(os.Stdout)
		w.Write([]string{"id", "weight", "created_at", "source", "notes"})
		w.Write([]string{
			fmt.Sprintf("%d", rowID),
			fmt.Sprintf("%.1f", weight),
			createdAt,
			source,
			notes,
		})
		w.Flush()
		return w.Error()
	default:
		fmt.Printf("Logged %.1f lb at %s (row id %d).\n", weight, createdAt, rowID)
		return nil
	}
}
