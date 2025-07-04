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
	Short: "Generate GORM models from database schema",
	Long:  "Connects to a database and generates GORM model files based on the existing table schemas.",
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.PrintInfo("✍️  Generating GORM models...", nil)

		dst, _ := cmd.Flags().GetString("dst")
		table, _ := cmd.Flags().GetString("table")
		exclude, _ := cmd.Flags().GetString("exclude")

		if table != "" && exclude != "" {
			return errors.New("flags --table and --exclude cannot be used at the same time")
		}

		generator := gen.NewModelGenerator(dst, table, exclude)
		if err := generator.Gen(); err != nil {
			if errors.Is(err, gen.ErrEnvFileNeedUpdate) {
				return nil
			}
			return err
		}

		utils.PrintSuccess("✅ Successfully generated GORM models.", nil)
		return nil
	},
}

func init() {
	genCmd.AddCommand(modelCmd)

	modelCmd.Flags().StringP("dst", "d", "internal/model", "Destination path for generated files")
	modelCmd.Flags().StringP("table", "t", "", "generate models for specific tables, separated by commas")
	modelCmd.Flags().StringP("exclude", "x", "", "exclude specific tables from generation, separated by commas")
}
