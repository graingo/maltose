package cli

import (
	"github.com/graingo/maltose/cmd/maltose/i18n"
	"github.com/spf13/cobra"
)

// genCmd represents the gen command which is a parent for other generation commands.
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: i18n.T("gen_cmd_short", nil),
	Long:  i18n.T("gen_cmd_long", nil),
}

func init() {
	rootCmd.AddCommand(genCmd)
}
