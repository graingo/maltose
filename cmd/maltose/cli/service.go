package cli

import (
	"fmt"

	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: utils.Print("service_cmd_short"),
	Long:  utils.Print("service_cmd_long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		src, _ := cmd.Flags().GetString("src")
		dst, _ := cmd.Flags().GetString("dst")
		mode, _ := cmd.Flags().GetString("mode")

		utils.PrintInfo("service_generation_start", nil)

		generator, err := gen.NewServiceGenerator(src, dst, mode == "interface")
		if err != nil {
			return err
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

	serviceCmd.Flags().StringP("src", "s", "api", "Source path for API definition files (directory or file)")
	serviceCmd.Flags().StringP("dst", "d", "internal", "Destination path for generated files")
	serviceCmd.Flags().StringP("mode", "m", "interface", "Generation mode: 'interface' or 'struct'")
}
