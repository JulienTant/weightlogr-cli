package cmd

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/julientant/weightlogr-cli/internal/db"
	"github.com/julientant/weightlogr-cli/internal/logger"
	"github.com/julientant/weightlogr-cli/internal/presentation"
	"github.com/julientant/weightlogr-cli/internal/store"
)

var updateCmd = &cobra.Command{
	Use:   "update <id> <weight>",
	Short: "Update a weigh-in",
	Args:  cobra.ExactArgs(2),
	RunE:  runUpdate,
}

func init() {
	updateCmd.Flags().String("source", DefaultSource, "Source label")
	updateCmd.Flags().String("notes", "", "Optional notes")

	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	log := logger.FromContext(ctx)

	log.Debug("parsing update arguments", "raw_id", args[0], "raw_weight", args[1])

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}

	weight, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid weight: %w", err)
	}
	if math.IsNaN(weight) || math.IsInf(weight, 0) {
		return fmt.Errorf("invalid weight: must be a finite number")
	}

	source, err := cmd.Flags().GetString("source")
	if err != nil {
		log.Warn("failed to read source flag", "error", err)
	}
	notes, err := cmd.Flags().GetString("notes")
	if err != nil {
		log.Warn("failed to read notes flag", "error", err)
	}

	log.Info("updating weigh-in", "id", id, "weight", weight, "source", source)

	conn, err := db.Open(ctx, viper.GetString("db"))
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer withLogError(ctx, conn.Close)

	s := store.New(conn)
	result, err := s.Update(ctx, id, weight, source, notes)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return fmt.Errorf("weigh-in %d not found: %w", id, err)
		}
		return fmt.Errorf("update weigh-in: %w", err)
	}

	return presentation.FormatUpdate(os.Stdout, viper.GetString("format"), result)
}
