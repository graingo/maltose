package cli

import (
	"os"
	"path/filepath"

	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

var (
	serviceSrcPath string
	serviceDstPath string
	serviceGenMode string
)

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service [path]",
	Short: utils.Print("service_cmd_short"),
	Long:  utils.Print("service_cmd_long"),
	Run: func(cmd *cobra.Command, args []string) {
		absSrc, err := filepath.Abs(serviceSrcPath)
		if err != nil {
			utils.PrintError("failed_to_get_abs_path", utils.TplData{"Error": err})
			os.Exit(1)
		}

		moduleName, moduleRoot, err := utils.GetModuleInfo(absSrc)
		if err != nil {
			utils.PrintError("go_mod_not_found", utils.TplData{"Error": err})
			os.Exit(1)
		}

		generator := &gen.ServiceGenerator{
			SrcPath:       absSrc,
			DstPath:       serviceDstPath,
			Module:        moduleName,
			ModuleRoot:    moduleRoot,
			InterfaceMode: serviceGenMode == "interface",
		}

		if err := generator.Gen(); err != nil {
			utils.PrintError("generic_error", utils.TplData{"Error": err})
			os.Exit(1)
		}

		utils.PrintSuccess("service_generation_success", nil)
	},
}

func init() {
	genCmd.AddCommand(serviceCmd)

	serviceCmd.Flags().StringVarP(&serviceSrcPath, "src", "s", "api", "Source path for API definition files (directory or file)")
	serviceCmd.Flags().StringVarP(&serviceDstPath, "dst", "d", "internal", "Destination path for generated files")
	serviceCmd.Flags().StringVarP(&serviceGenMode, "mode", "m", "interface", "Generation mode: 'interface' or 'struct'")
}
