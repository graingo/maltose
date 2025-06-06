package cli

import (
	"os"
	"path/filepath"

	"github.com/graingo/maltose/cmd/maltose/i18n"
	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

// logicCmd represents the logic command
var logicCmd = &cobra.Command{
	Use:   "logic [path]",
	Short: i18n.T("logic_cmd_short", nil),
	Long:  i18n.T("logic_cmd_long", nil),
	Run: func(cmd *cobra.Command, args []string) {
		// Priority: argument > flag
		if len(args) > 0 {
			srcPath = args[0]
		}

		absSrc, err := filepath.Abs(srcPath)
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
			Module:     moduleName,
			ModuleRoot: moduleRoot,
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

	logicCmd.Flags().StringVarP(&srcPath, "src", "s", "internal/service", "Source path for service definition files")
}
