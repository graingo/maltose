package cli

import (
	"fmt"

	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

var (
	daoSrc string
)

// daoCmd represents the dao command
var daoCmd = &cobra.Command{
	Use:   "dao",
	Short: utils.Print("dao_cmd_short"),
	Long:  utils.Print("dao_cmd_long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.PrintInfo("dao_generation_start", nil)

		modulePath, _, err := utils.GetModuleInfo(".")
		if err != nil {
			return fmt.Errorf("could not find go.mod: %w", err)
		}

		generator := gen.NewDaoGenerator(modulePath)
		if err := generator.Gen(); err != nil {
			return err
		}

		utils.PrintSuccess("dao_generation_success", nil)
		return nil
	},
}

func init() {
	genCmd.AddCommand(daoCmd)

	daoCmd.Flags().StringVarP(&daoSrc, "src", "s", "internal/model/entity", "Source path for model entity files")
}
