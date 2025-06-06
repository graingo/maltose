package cli

import (
	"fmt"

	"github.com/graingo/maltose/cmd/maltose/i18n"
	"github.com/spf13/cobra"
)

var (
	// Version is the current version of the application.
	Version = "0.0.1"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: i18n.T("version_cmd_short", nil),
	Long:  i18n.T("version_cmd_long", nil),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Maltose version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
