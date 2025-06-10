package cli

import (
	"fmt"

	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

// logicCmd represents the logic command
var logicCmd = &cobra.Command{
	Use:   "logic [path]",
	Short: utils.Print("logic_cmd_short"),
	Long:  utils.Print("logic_cmd_long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.PrintInfo("logic_generation_start", nil)

		srcPath, _ := cmd.Flags().GetString("src")
		dstPath, _ := cmd.Flags().GetString("dst")
		overwrite, _ := cmd.Flags().GetBool("overwrite")

		generator, err := gen.NewLogicGenerator(srcPath, dstPath, overwrite)
		if err != nil {
			return err
		}
		if err := generator.Gen(); err != nil {
			return fmt.Errorf("failed to generate logic file: %w", err)
		}

		utils.PrintSuccess("logic_generation_success", nil)
		return nil
	},
}

func init() {
	genCmd.AddCommand(logicCmd)

	logicCmd.Flags().StringP("src", "s", "internal/service", "Source path for service definition files")
	logicCmd.Flags().StringP("dst", "d", "internal", "Destination path for generated files")
	logicCmd.Flags().BoolP("overwrite", "o", false, "Overwrite existing logic file if it exists")
}
