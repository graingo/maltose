package cli

import (
	"os"

	"github.com/graingo/maltose/cmd/maltose/i18n"
	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

// daoCmd represents the dao command
var daoCmd = &cobra.Command{
	Use:   "dao",
	Short: i18n.T("dao_cmd_short", nil),
	Long:  i18n.T("dao_cmd_long", nil),
	Run: func(cmd *cobra.Command, args []string) {
		utils.PrintInfo("dao_generation_start", nil)

		modulePath, _, err := utils.GetModuleInfo(".")
		if err != nil {
			utils.PrintError("generic_error", utils.TplData{"Error": err})
			os.Exit(1)
		}

		generator := gen.NewDaoGenerator(modulePath)
		if err := generator.Gen(); err != nil {
			utils.PrintError("generic_error", utils.TplData{"Error": err})
			os.Exit(1)
		}

		utils.PrintSuccess("dao_generation_success", nil)
	},
}

func init() {
	genCmd.AddCommand(daoCmd)
}
