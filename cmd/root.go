package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "weightlogr",
	Short: "Weight tracking CLI optimized for AI agents",
	Long:  "A 12-factor weight tracking CLI. Configure via flags, env vars (WEIGHTLOGR_ prefix), or config file.",
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		// 1. Initialize Viper: config file paths, env prefix, read config.
		viper.SetConfigName(".weightlogr")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("$HOME/.config/weightlogr")
		viper.AddConfigPath("/etc/weightlogr")
		viper.AddConfigPath(".")
		viper.SetEnvPrefix("WEIGHTLOGR")
		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return fmt.Errorf("read config: %w", err)
			}
		}

		// 3. Logger setup — values read from Viper (single source of truth).
		logger, err := setupLogger()
		if err != nil {
			return err
		}
		ctx := contextWithLogger(cmd.Context(), logger)
		cmd.SetContext(ctx)

		if f := viper.ConfigFileUsed(); f != "" {
			logger.Debug("config file loaded", "path", f)
		}

		logger.Info("weightlogr starting", "command", cmd.Name())
		logger.Debug("resolved configuration",
			"db", viper.GetString("db"),
			"timezone", viper.GetString("timezone"),
			"format", viper.GetString("format"),
			"log_file", viper.GetString("log_file"),
			"log_level", viper.GetString("log_level"),
		)
		return nil
	},
}

func Execute() {
	ctx := context.Background()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// 2. Define flags and bind to Viper in init().
func init() {
	f := rootCmd.PersistentFlags()

	f.String("db", "/opt/data/weights.db", "Path to SQLite database")
	f.String("timezone", "America/Phoenix", "Timezone for timestamps")
	f.String("format", "table", "Output format: table, json, csv")
	f.String("log-file", "/opt/data/weightlogr.log", "Path to log file (use 'stderr' for stderr)")
	f.String("log-level", "info", "Log level: debug, info, warn, error")

	viper.BindPFlag("db", f.Lookup("db"))
	viper.BindPFlag("timezone", f.Lookup("timezone"))
	viper.BindPFlag("format", f.Lookup("format"))
	viper.BindPFlag("log_file", f.Lookup("log-file"))
	viper.BindPFlag("log_level", f.Lookup("log-level"))
}

// 3. All helpers below read from Viper, never from flag variables directly.

func setupLogger() (*slog.Logger, error) {
	logFile := viper.GetString("log_file")
	level := parseLogLevel(viper.GetString("log_level"))

	var handler slog.Handler
	if logFile == "stderr" {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	} else {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("open log file %q: %w", logFile, err)
		}
		handler = slog.NewJSONHandler(f, &slog.HandlerOptions{Level: level})
	}

	return slog.New(handler), nil
}

func parseLogLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type ctxKeyLogger struct{}

func contextWithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger{}, l)
}

func loggerFrom(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKeyLogger{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
