package cli

import (
	"os"
	"path/filepath"

	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/spf13/cobra"
)

var logicCmd = &cobra.Command{
	Use:   "logic",
	Short: "Generate logic file from service interface",
	Long:  `Generate logic file implementation from a service interface file.`,
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

	logicCmd.Flags().StringVarP(&srcPath, "src", "s", "internal/service", "Source service interface file (required)")
	logicCmd.Flags().StringVarP(&dstPath, "dst", "d", "internal/logic", "Destination directory for generated files")
}
