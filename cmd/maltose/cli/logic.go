package cli

import (
	"fmt"
	"path/filepath"

	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

var (
	logicSrcPath   string
	logicDstPath   string
	logicOverwrite bool
)

// logicCmd represents the logic command
var logicCmd = &cobra.Command{
	Use:   "logic [path]",
	Short: utils.Print("logic_cmd_short"),
	Long:  utils.Print("logic_cmd_long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		absSrc, err := filepath.Abs(logicSrcPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute source path: %w", err)
		}

		moduleName, moduleRoot, err := utils.GetModuleInfo(absSrc)
		if err != nil {
			return fmt.Errorf("could not find go.mod: %w", err)
		}

		generator := &gen.LogicGenerator{
			SrcPath:    absSrc,
			DstPath:    logicDstPath,
			Module:     moduleName,
			ModuleRoot: moduleRoot,
			Overwrite:  logicOverwrite,
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

	logicCmd.Flags().StringVarP(&logicSrcPath, "src", "s", "internal/service", "Source path for service definition files")
	logicCmd.Flags().StringVarP(&logicDstPath, "dst", "d", "internal", "Destination path for generated files")
	logicCmd.Flags().BoolVar(&logicOverwrite, "overwrite", false, "Overwrite existing logic file if it exists")
}
