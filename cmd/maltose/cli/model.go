package cli

import (
	"os"

	"github.com/graingo/maltose/cmd/maltose/i18n"
	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

// modelCmd represents the model command
var modelCmd = &cobra.Command{
	Use:   "model",
	Short: i18n.T("model_cmd_short", nil),
	Long:  i18n.T("model_cmd_long", nil),
	Run: func(cmd *cobra.Command, args []string) {
		utils.PrintInfo("gorm_model_generation_start", nil)

		generator := gen.NewModelGenerator()
		if err := generator.Gen(); err != nil {
			utils.PrintError("generic_error", utils.TplData{"Error": err})
			os.Exit(1)
		}

		utils.PrintSuccess("gorm_model_generation_success", nil)
	},
}

func init() {
	genCmd.AddCommand(modelCmd)
}
