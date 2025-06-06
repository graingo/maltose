package cli

import (
	"os"
	"path/filepath"

	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

// logicCmd represents the logic command
var logicCmd = &cobra.Command{
	Use:   "logic [path]",
	Short: "Generate logic file from service definitions",
	Long:  `Generate logic file based on Go files containing service interface definitions.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Priority: argument > flag
		if len(args) > 0 {
			srcPath = args[0]
		}

		absSrc, err := filepath.Abs(srcPath)
		if err != nil {
			utils.PrintError("failedToGetAbsPath", map[string]interface{}{"Error": err})
			os.Exit(1)
		}

		moduleName, moduleRoot, err := findModuleInfo(absSrc)
		if err != nil {
			utils.PrintError("goModNotFound", map[string]interface{}{"Error": err})
			os.Exit(1)
		}

		generator := &gen.LogicGenerator{
			SrcPath:    absSrc,
			Module:     moduleName,
			ModuleRoot: moduleRoot,
		}

		if err := generator.Gen(); err != nil {
			utils.PrintError("logicGenerationFailed", map[string]interface{}{"Error": err})
			os.Exit(1)
		}

		utils.PrintSuccess("logicGenerationSuccess", nil)
	},
}

func init() {
	genCmd.AddCommand(logicCmd)

	logicCmd.Flags().StringVarP(&srcPath, "src", "s", "internal/service", "Source path for service definition files")
}
