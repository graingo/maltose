package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
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
			utils.PrintError("failedToGetAbsPath", map[string]interface{}{"Error": err})
			os.Exit(1)
		}

		moduleName, moduleRoot, err := findModuleInfo(absSrc)
		if err != nil {
			utils.PrintError("goModNotFound", map[string]interface{}{"Error": err})
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
			utils.PrintError("genericError", map[string]interface{}{"Error": err})
			os.Exit(1)
		}

		utils.PrintSuccess("serviceGenerationSuccess", nil)
	},
}

func findModuleInfo(fromPath string) (name, rootPath string, err error) {
	currentPath := fromPath
	for {
		goModPath := filepath.Join(currentPath, "go.mod")
		content, err := os.ReadFile(goModPath)
		if err == nil {
			return modfile.ModulePath(content), currentPath, nil
		}

		parent := filepath.Dir(currentPath)
		if parent == currentPath { // Reached the root
			return "", "", fmt.Errorf("go.mod not found")
		}
		currentPath = parent
	}
}

func init() {
	genCmd.AddCommand(serviceCmd)

	serviceCmd.Flags().StringVarP(&srcPath, "src", "s", "api", "Source path for API definition files (directory or file)")
	serviceCmd.Flags().StringVarP(&dstPath, "dst", "d", "internal", "Destination path for generated files")
	serviceCmd.Flags().StringVarP(&genMode, "mode", "m", "interface", "Generation mode: 'interface' or 'struct'")
}
