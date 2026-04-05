package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/julientant/weightlogr-cli/internal/logger"
	"github.com/julientant/weightlogr-cli/internal/presentation"
	"github.com/julientant/weightlogr-cli/internal/version"
)

const (
	DefaultDB       = "/opt/data/weights.db"
	DefaultLogFile  = "/opt/data/weightlogr.log"
	DefaultLogLevel = "info"

	LogFileStderr = "stderr"

	LogFilePermissions = 0o644

	LogLevelDebug = "debug"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
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
		l, err := setupLogger()
		if err != nil {
			return fmt.Errorf("setup logger: %w", err)
		}
		ctx := logger.WithContext(cmd.Context(), l)
		cmd.SetContext(ctx)

		if f := viper.ConfigFileUsed(); f != "" {
			l.Debug("config file loaded", "path", f)
		}

		buildInfo := version.Info()
		l.Debug("build info",
			"version", buildInfo.Version,
			"commit", buildInfo.Commit,
			"date", buildInfo.Date,
		)
		l.Info("weightlogr starting", "command", cmd.Name())
		l.Debug("resolved configuration",
			"db", viper.GetString("db"),
			"format", viper.GetString("format"),
			"log_file", viper.GetString("log_file"),
			"log_level", viper.GetString("log_level"),
		)
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// 2. Define flags and bind to Viper in init().
func init() {
	f := rootCmd.PersistentFlags()

	f.String("db", DefaultDB, "Path to SQLite database")
	f.String("format", presentation.FormatJSON, "Output format: json, csv")
	f.String("log-file", DefaultLogFile, "Path to log file (use 'stderr' for stderr)")
	f.String("log-level", DefaultLogLevel, "Log level: debug, info, warn, error")

	for _, b := range []struct {
		key  string
		flag string
	}{
		{"db", "db"},
		{"format", "format"},
		{"log_file", "log-file"},
		{"log_level", "log-level"},
	} {
		if err := viper.BindPFlag(b.key, f.Lookup(b.flag)); err != nil {
			panic(fmt.Sprintf("bind flag %q: %v", b.flag, err))
		}
	}
}

// 3. All helpers below read from Viper, never from flag variables directly.

func setupLogger() (*slog.Logger, error) {
	logFile := viper.GetString("log_file")
	level := parseLogLevel(viper.GetString("log_level"))

	var handler slog.Handler
	if logFile == LogFileStderr {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	} else {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, LogFilePermissions)
		if err != nil {
			return nil, fmt.Errorf("open log file %q: %w", logFile, err)
		}
		handler = slog.NewJSONHandler(f, &slog.HandlerOptions{Level: level})
	}

	return slog.New(handler), nil
}

func parseLogLevel(s string) slog.Level {
	switch s {
	case LogLevelDebug:
		return slog.LevelDebug
	case LogLevelWarn:
		return slog.LevelWarn
	case LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
