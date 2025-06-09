package cli

import (
	"fmt"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		absSrc, err := filepath.Abs(serviceSrcPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute source path: %w", err)
		}

		moduleName, moduleRoot, err := utils.GetModuleInfo(absSrc)
		if err != nil {
			return fmt.Errorf("could not find go.mod: %w", err)
		}

		generator := &gen.ServiceGenerator{
			SrcPath:       absSrc,
			DstPath:       serviceDstPath,
			Module:        moduleName,
			ModuleRoot:    moduleRoot,
			InterfaceMode: serviceGenMode == "interface",
		}

		if err := generator.Gen(); err != nil {
			return fmt.Errorf("failed to generate service file: %w", err)
		}

		utils.PrintSuccess("service_generation_success", nil)
		return nil
	},
}

func init() {
	genCmd.AddCommand(serviceCmd)

	serviceCmd.Flags().StringVarP(&serviceSrcPath, "src", "s", "api", "Source path for API definition files (directory or file)")
	serviceCmd.Flags().StringVarP(&serviceDstPath, "dst", "d", "internal", "Destination path for generated files")
	serviceCmd.Flags().StringVarP(&serviceGenMode, "mode", "m", "interface", "Generation mode: 'interface' or 'struct'")
}
