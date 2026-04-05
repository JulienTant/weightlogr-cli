package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/julientant/weightlogr-cli/internal/db"
	"github.com/julientant/weightlogr-cli/internal/logger"
	"github.com/julientant/weightlogr-cli/internal/presentation"
	"github.com/julientant/weightlogr-cli/internal/store"
)

const DefaultSource = "daily-check"

var insertCmd = &cobra.Command{
	Use:   "insert <weight>",
	Short: "Log a weigh-in",
	Args:  cobra.ExactArgs(1),
	RunE:  runInsert,
}

func init() {
	insertCmd.Flags().String("timestamp", "", "RFC3339 timestamp (defaults to now in UTC)")
	insertCmd.Flags().String("source", DefaultSource, "Source label")
	insertCmd.Flags().String("notes", "", "Optional notes")

	rootCmd.AddCommand(insertCmd)
}

func runInsert(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	log := logger.FromContext(ctx)

	log.Debug("parsing insert arguments", "raw_weight", args[0])

	var weight float64
	if _, err := fmt.Sscanf(args[0], "%f", &weight); err != nil {
		return fmt.Errorf("invalid weight: %w", err)
	}

	ts, err := cmd.Flags().GetString("timestamp")
	if err != nil {
		log.Warn("failed to read timestamp flag", "error", err)
	}
	createdAt, err := normalizeTimestamp(ts)
	if err != nil {
		return fmt.Errorf("normalize timestamp: %w", err)
	}
	source, err := cmd.Flags().GetString("source")
	if err != nil {
		log.Warn("failed to read source flag", "error", err)
	}
	notes, err := cmd.Flags().GetString("notes")
	if err != nil {
		log.Warn("failed to read notes flag", "error", err)
	}

	log.Info("inserting weigh-in", "weight", weight, "created_at", createdAt, "source", source)

	conn, err := db.Open(ctx, viper.GetString("db"))
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer withLogError(ctx, conn.Close)

	s := store.New(conn)
	result, err := s.Insert(ctx, weight, createdAt, source, notes)
	if err != nil {
		return fmt.Errorf("insert weigh-in: %w", err)
	}

	return presentation.FormatInsert(os.Stdout, viper.GetString("format"), result)
}
