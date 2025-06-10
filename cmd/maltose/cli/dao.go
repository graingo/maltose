package cli

import (
	"errors"

	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

// daoCmd represents the dao command
var daoCmd = &cobra.Command{
	Use:   "dao",
	Short: utils.Print("dao_cmd_short"),
	Long:  utils.Print("dao_cmd_long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.PrintInfo("dao_generation_start", nil)

		dst, _ := cmd.Flags().GetString("dst")

		generator, err := gen.NewDaoGenerator(dst)
		if err != nil {
			return err
		}
		if err := generator.Gen(); err != nil {
			if errors.Is(err, gen.ErrEnvFileNeedUpdate) {
				return nil
			}
			return err
		}

		utils.PrintSuccess("dao_generation_success", nil)
		return nil
	},
}

func init() {
	genCmd.AddCommand(daoCmd)

	daoCmd.Flags().StringP("dst", "d", "internal/dao", "Destination path for generated files")
}
