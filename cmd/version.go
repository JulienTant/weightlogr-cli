package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/julientant/weightlogr-cli/internal/presentation"
	"github.com/julientant/weightlogr-cli/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	RunE:  runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(_ *cobra.Command, _ []string) error {
	return presentation.FormatVersion(os.Stdout, viper.GetString("format"), version.Info())
}
