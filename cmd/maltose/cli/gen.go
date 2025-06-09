package cli

import (
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

// genCmd represents the gen command which is a parent for other generation commands.
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: utils.Print("gen_cmd_short"),
	Long:  utils.Print("gen_cmd_long"),
}

func init() {
	rootCmd.AddCommand(genCmd)
}
