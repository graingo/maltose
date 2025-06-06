package cli

import (
	"os"
	"path/filepath"

	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

var (
	srcPath string
	dstPath string
	genMode string
)

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service [path]",
	Short: "Generate controller and service files from api definitions",
	Long: `Generate controller and service files based on Go files containing API
definitions (structs for request and response).

The command defaults to using 'api' as input and 'internal' as output.
You can provide a single file or a directory as input. When a directory is provided,
it will recursively find all .go files.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Priority: argument > flag
		if len(args) > 0 {
			srcPath = args[0]
		}

		absSrc, err := filepath.Abs(srcPath)
		if err != nil {
			utils.PrintError("failedToGetAbsPath", utils.TplData{"Error": err})
			os.Exit(1)
		}

		moduleName, moduleRoot, err := utils.GetModuleInfo(absSrc)
		if err != nil {
			utils.PrintError("goModNotFound", utils.TplData{"Error": err})
			os.Exit(1)
		}

		generator := &gen.ServiceGenerator{
			SrcPath:       absSrc,
			DstPath:       dstPath,
			Module:        moduleName,
			ModuleRoot:    moduleRoot,
			InterfaceMode: genMode == "interface",
		}

		if err := generator.Gen(); err != nil {
			utils.PrintError("genericError", utils.TplData{"Error": err})
			os.Exit(1)
		}

		utils.PrintSuccess("serviceGenerationSuccess", nil)
	},
}

func init() {
	genCmd.AddCommand(serviceCmd)

	serviceCmd.Flags().StringVarP(&srcPath, "src", "s", "api", "Source path for API definition files (directory or file)")
	serviceCmd.Flags().StringVarP(&dstPath, "dst", "d", "internal", "Destination path for generated files")
	serviceCmd.Flags().StringVarP(&genMode, "mode", "m", "interface", "Generation mode: 'interface' or 'struct'")
}
