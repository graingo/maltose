package cli

import (
	"fmt"

	"github.com/graingo/maltose"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Maltose",
	Long:  `All software has versions. This is Maltose's.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Maltose CLI version: %s\n", maltose.VERSION)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
