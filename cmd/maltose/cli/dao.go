package cli

import (
	"os"

	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

// daoCmd represents the dao command
var daoCmd = &cobra.Command{
	Use:   "dao",
	Short: "Generate DAO layer based on existing models.",
	Long: `This command scans for GORM models and generates a complete data access object (DAO) layer, including interfaces and implementations.
`,
	Run: func(cmd *cobra.Command, args []string) {
		utils.PrintInfo("daoGenerationStart", nil)

		modulePath, _, err := utils.GetModuleInfo(".")
		if err != nil {
			utils.PrintError("genericError", utils.TplData{"Error": err})
			os.Exit(1)
		}

		generator := gen.NewDaoGenerator(modulePath)
		if err := generator.Gen(); err != nil {
			utils.PrintError("genericError", utils.TplData{"Error": err})
			os.Exit(1)
		}

		utils.PrintSuccess("daoGenerationSuccess", nil)
	},
}

func init() {
	genCmd.AddCommand(daoCmd)
}
