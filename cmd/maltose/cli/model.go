package cli

import (
	"errors"

	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

// modelCmd represents the model command
var modelCmd = &cobra.Command{
	Use:   "model",
	Short: utils.Print("model_cmd_short"),
	Long:  utils.Print("model_cmd_long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.PrintInfo("gorm_model_generation_start", nil)

		generator := gen.NewModelGenerator()
		if err := generator.Gen(); err != nil {
			if errors.Is(err, gen.ErrEnvFileNeedUpdate) {
				return nil
			}
			return err
		}

		utils.PrintSuccess("gorm_model_generation_success", nil)
		return nil
	},
}

func init() {
	genCmd.AddCommand(modelCmd)
}
