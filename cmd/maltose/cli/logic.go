package cli

import (
	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/graingo/maltose/errors/merror"
	"github.com/spf13/cobra"
)

// logicCmd represents the logic command
var logicCmd = &cobra.Command{
	Use:   "logic [path]",
	Short: "Generate logic file from service definitions",
	Long:  "Generate logic file based on Go files containing service interface definitions.",
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.PrintInfo("✍️  Generating logic files...", nil)

		srcPath, _ := cmd.Flags().GetString("src")
		dstPath, _ := cmd.Flags().GetString("dst")
		overwrite, _ := cmd.Flags().GetBool("overwrite")

		generator, err := gen.NewLogicGenerator(srcPath, dstPath, overwrite)
		if err != nil {
			return err
		}
		if err := generator.Gen(); err != nil {
			return merror.Wrap(err, "failed to generate logic file")
		}

		utils.PrintSuccess("✅ Successfully generated logic files.", nil)
		return nil
	},
}

func init() {
	genCmd.AddCommand(logicCmd)

	logicCmd.Flags().StringP("src", "s", "internal/service", "Source path for service definition files")
	logicCmd.Flags().StringP("dst", "d", "internal", "Destination path for generated files")
	logicCmd.Flags().BoolP("overwrite", "o", false, "Overwrite existing logic file if it exists")
}
