package cli

import (
	"os"
	"path/filepath"

	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/spf13/cobra"
)

var logicCmd = &cobra.Command{
	Use:   "logic [path]",
	Short: "Generate logic files from service interfaces",
	Long:  `Generate logic file implementations from service interface files.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Priority: argument > flag
		if len(args) > 0 {
			srcPath = args[0]
		}

		absSrc, err := filepath.Abs(srcPath)
		if err != nil {
			PrintError("Failed to get absolute source path: %v\n", err)
			os.Exit(1)
		}

		moduleName, moduleRoot, err := findModuleInfo(absSrc)
		if err != nil {
			PrintError("Could not find go.mod to determine module info: %v. Please run this command in a valid Go module.\n", err)
			os.Exit(1)
		}

		generator := &gen.LogicGenerator{
			SrcPath:    srcPath,
			DstPath:    dstPath,
			Module:     moduleName,
			ModuleRoot: moduleRoot,
		}

		if err := generator.Gen(); err != nil {
			PrintError("Failed to generate logic file: %v\n", err)
			os.Exit(1)
		}

		PrintSuccess("Logic file generated successfully.\n")
	},
}

func init() {
	genCmd.AddCommand(logicCmd)

	logicCmd.Flags().StringVarP(&srcPath, "src", "s", "internal/service", "Source directory for service interface files")
	logicCmd.Flags().StringVarP(&dstPath, "dst", "d", "internal", "Destination root directory for generated files")
}
