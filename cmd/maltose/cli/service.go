package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
)

var (
	srcPath    string
	dstPath    string
	moduleName string
	moduleRoot string
	genMode    string
)

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service [path]",
	Short: "Generate controller and service files from api definitions",
	Long: `Generate controller and service files based on Go files containing API
definitions (structs for request and response).

You can provide a single file or a directory as input. When a directory is provided,
it will recursively find all .go files.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			srcPath = args[0]
		}
		if srcPath == "" {
			srcPath, _ = os.Getwd()
		}

		if dstPath == "" {
			dstPath, _ = os.Getwd()
		}

		absSrc, err := filepath.Abs(srcPath)
		if err != nil {
			log.Fatalf("Error getting absolute source path: %v", err)
		}

		if moduleName == "" || moduleRoot == "" {
			var err error
			moduleName, moduleRoot, err = findModuleInfo(absSrc)
			if err != nil {
				log.Printf("Warning: Could not find go.mod to determine module info: %v. Import paths might be incorrect.", err)
			}
		}

		generator := &gen.ServiceGenerator{
			SrcPath:       absSrc,
			DstPath:       dstPath,
			Module:        moduleName,
			ModuleRoot:    moduleRoot,
			InterfaceMode: genMode == "interface",
		}

		if err := generator.Gen(); err != nil {
			log.Fatalf("Error generating services: %v", err)
		}

		fmt.Println("Service and controller files generated successfully.")
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

	serviceCmd.Flags().StringVarP(&srcPath, "src", "s", "", "Source path for API definition files (directory or file)")
	serviceCmd.Flags().StringVarP(&dstPath, "dst", "d", "internal", "Destination path for generated files")
	serviceCmd.Flags().StringVarP(&moduleName, "module", "m", "", "Go module name of the project (e.g., github.com/user/project)")
	serviceCmd.Flags().StringVar(&genMode, "mode", "interface", "Generation mode: 'interface' or 'struct'")
}
