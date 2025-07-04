package cli

import (
	"github.com/spf13/cobra"
)

// genCmd represents the gen command which is a parent for other generation commands.
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate various codes (service, model, etc.)",
	Long:  "A collection of code generation commands.",
}

func init() {
	rootCmd.AddCommand(genCmd)
}
