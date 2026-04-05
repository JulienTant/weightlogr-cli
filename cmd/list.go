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

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List weigh-ins",
	RunE:  runList,
}

func init() {
	listCmd.Flags().String("since", "", "Start date/time inclusive (ISO 8601)")
	listCmd.Flags().String("until", "", "End date/time exclusive (ISO 8601)")
	listCmd.Flags().String("source", "", "Filter by source")
	listCmd.Flags().String("order", store.OrderDesc, "Sort order: asc or desc")
	listCmd.Flags().Int("limit", 0, "Max rows to return (0 = unlimited)")

	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	log := logger.FromContext(ctx)

	log.Debug("parsing list flags")

	tz, err := loadTimezone(ctx)
	if err != nil {
		return fmt.Errorf("load timezone: %w", err)
	}

	conn, err := db.Open(ctx, viper.GetString("db"))
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer withLogError(ctx, conn.Close)

	since, err := cmd.Flags().GetString("since")
	if err != nil {
		log.Warn("failed to read since flag", "error", err)
	}
	until, err := cmd.Flags().GetString("until")
	if err != nil {
		log.Warn("failed to read until flag", "error", err)
	}
	if since != "" {
		since, err = normalizeTimestamp(ctx, since, tz)
		if err != nil {
			return fmt.Errorf("normalize since timestamp: %w", err)
		}
	}
	if until != "" {
		until, err = normalizeTimestamp(ctx, until, tz)
		if err != nil {
			return fmt.Errorf("normalize until timestamp: %w", err)
		}
	}

	source, err := cmd.Flags().GetString("source")
	if err != nil {
		log.Warn("failed to read source flag", "error", err)
	}
	order, err := cmd.Flags().GetString("order")
	if err != nil {
		log.Warn("failed to read order flag", "error", err)
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		log.Warn("failed to read limit flag", "error", err)
	}

	s := store.New(conn)
	results, err := s.List(ctx, store.ListOpts{
		Since:  since,
		Until:  until,
		Source: source,
		Order:  order,
		Limit:  limit,
	})
	if err != nil {
		return fmt.Errorf("list weigh-ins: %w", err)
	}

	return presentation.FormatList(os.Stdout, viper.GetString("format"), viper.GetString("unit"), results)
}
