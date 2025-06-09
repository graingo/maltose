package cli

import (
	"os"
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
	Run: func(cmd *cobra.Command, args []string) {
		absSrc, err := filepath.Abs(logicSrcPath)
		if err != nil {
			utils.PrintError("failed_to_get_abs_path", utils.TplData{"Error": err})
			os.Exit(1)
		}

		moduleName, moduleRoot, err := utils.GetModuleInfo(absSrc)
		if err != nil {
			utils.PrintError("go_mod_not_found", utils.TplData{"Error": err})
			os.Exit(1)
		}

		generator := &gen.LogicGenerator{
			SrcPath:    absSrc,
			DstPath:    logicDstPath,
			Module:     moduleName,
			ModuleRoot: moduleRoot,
			Overwrite:  logicOverwrite,
		}

		if err := generator.Gen(); err != nil {
			utils.PrintError("logic_generation_failed", utils.TplData{"Error": err})
			os.Exit(1)
		}

		utils.PrintSuccess("logic_generation_success", nil)
	},
}

func init() {
	genCmd.AddCommand(logicCmd)

	logicCmd.Flags().StringVarP(&logicSrcPath, "src", "s", "internal/service", "Source path for service definition files")
	logicCmd.Flags().StringVarP(&logicDstPath, "dst", "d", "internal", "Destination path for generated files")
	logicCmd.Flags().BoolVar(&logicOverwrite, "overwrite", false, "Overwrite existing logic file if it exists")
}
