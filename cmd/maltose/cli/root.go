package cli

import (
	"fmt"
	"os"

	"github.com/graingo/maltose"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

var (
	versionFlag bool
)

var rootCmd = &cobra.Command{
	Use:   "maltose",
	Short: utils.Print("root_cmd_short"),
	Long:  utils.Print("root_cmd_long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		if versionFlag {
			fmt.Printf("Maltose CLI version: %s\n", maltose.VERSION)
			return nil
		}
		// Default action when no subcommand is provided
		return cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = maltose.VERSION
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
