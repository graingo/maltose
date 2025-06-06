package cmd

import (
	"github.com/spf13/cobra"
)

// genCmd represents the gen command which is a parent for other generation commands.
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "A collection of code generation commands",
	Long:  `A collection of code generation commands for Maltose projects, such as generating DAO layers.`,
}

func init() {
	rootCmd.AddCommand(genCmd)
}
