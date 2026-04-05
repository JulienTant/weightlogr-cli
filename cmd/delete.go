package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/julientant/weightlogr-cli/internal/db"
	"github.com/julientant/weightlogr-cli/internal/logger"
	"github.com/julientant/weightlogr-cli/internal/presentation"
	"github.com/julientant/weightlogr-cli/internal/store"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a weigh-in (soft delete)",
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	log := logger.FromContext(ctx)

	log.Debug("parsing delete arguments", "raw_id", args[0])

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}

	log.Info("deleting weigh-in", "id", id)

	conn, err := db.Open(ctx, viper.GetString("db"))
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer withLogError(ctx, conn.Close)

	s := store.New(conn)
	if err := s.Delete(ctx, id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return fmt.Errorf("weigh-in %d not found: %w", id, err)
		}
		return fmt.Errorf("delete weigh-in: %w", err)
	}

	return presentation.FormatDelete(os.Stdout, viper.GetString("format"), id)
}
