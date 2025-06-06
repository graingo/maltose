package cli

import (
	"os"

	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

// modelCmd represents the model command
var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Generate GORM models from database schema",
	Long:  `Connects to a database and generates GORM model files based on the existing table schemas.`,
	Run: func(cmd *cobra.Command, args []string) {
		utils.PrintInfo("gormModelGenerationStart", nil)

		generator := gen.NewModelGenerator()
		if err := generator.Gen(); err != nil {
			utils.PrintError("genericError", utils.TplData{"Error": err})
			os.Exit(1)
		}

		utils.PrintSuccess("gormModelGenerationSuccess", nil)
	},
}

func init() {
	genCmd.AddCommand(modelCmd)
}
