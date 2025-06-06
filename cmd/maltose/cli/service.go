package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
)

var (
	srcPath     string
	dstPath     string
	moduleName  string
	modRootFlag string
	genMode     string
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

		var modRoot string

		// If --mod-root is set, use it directly.
		if modRootFlag != "" {
			var err error
			modRoot, err = filepath.Abs(modRootFlag)
			if err != nil {
				PrintError("Failed to get absolute path for --mod-root: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Otherwise, find it based on current working directory.
			cwd, err := os.Getwd()
			if err != nil {
				PrintError("Failed to get current working directory: %v\n", err)
				os.Exit(1)
			}
			_, modRoot, err = findModuleInfo(cwd)
			if err != nil {
				PrintError("Could not find go.mod to determine module info: %v. Please run from within a Go module or use the --mod-root flag.\n", err)
				os.Exit(1)
			}
		}

		// Read the go.mod file from the determined root to get the module name
		goModPath := filepath.Join(modRoot, "go.mod")
		content, err := os.ReadFile(goModPath)
		if err != nil {
			PrintError("Could not read go.mod at %s: %v\n", goModPath, err)
			os.Exit(1)
		}
		detectedModule := modfile.ModulePath(content)

		// Priority for module name: flag > detected
		if moduleName == "" {
			moduleName = detectedModule
		}

		// Source path needs to be absolute for reliable processing
		absSrc, err := filepath.Abs(srcPath)
		if err != nil {
			PrintError("Failed to get absolute source path: %v\n", err)
			os.Exit(1)
		}

		generator := &gen.ServiceGenerator{
			SrcPath:       absSrc,
			DstPath:       dstPath,
			Module:        moduleName,
			ModuleRoot:    modRoot,
			InterfaceMode: genMode == "interface",
		}

		if err := generator.Gen(); err != nil {
			PrintError("%v\n", err)
			os.Exit(1)
		}

		PrintSuccess("Service and controller files generated successfully.\n")
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
	serviceCmd.Flags().StringVar(&moduleName, "module", "", "Go module name (e.g., github.com/user/project). If not set, it is detected from go.mod.")
	serviceCmd.Flags().StringVar(&modRootFlag, "mod-root", "", "Manually specify the module root directory containing go.mod.")
}
