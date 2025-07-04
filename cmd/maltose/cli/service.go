package cli

import (
	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/graingo/maltose/errors/merror"
	"github.com/spf13/cobra"
)

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Generate controller and service files from api definitions",
	Long: `Generate controller and service files based on Go files containing API
definitions (structs for request and response).

The command defaults to using 'api' as input and 'internal' as output.
You can provide a single file or a directory as input. When a directory is provided,
it will recursively find all .go files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		src, _ := cmd.Flags().GetString("src")
		dst, _ := cmd.Flags().GetString("dst")
		mode, _ := cmd.Flags().GetString("mode")

		utils.PrintInfo("✍️  Generating controller and service files...", nil)

		generator, err := gen.NewServiceGenerator(src, dst, name, mode == "interface")
		if err != nil {
			return err
		}
		if err = generator.Gen(); err != nil {
			return merror.Wrap(err, "failed to generate service file")
		}

		utils.PrintSuccess("✅ Successfully generated controller and service files.", nil)
		return nil
	},
}

func init() {
	genCmd.AddCommand(serviceCmd)

	serviceCmd.Flags().StringP("name", "n", "", "Name of the service to generate (e.g., 'user'). This will create a single service interface file.")
	serviceCmd.Flags().StringP("src", "s", "api", "Source path for API definition files (directory or file). Ignored if --name is used.")
	serviceCmd.Flags().StringP("dst", "d", "internal", "Destination path for generated files")
	serviceCmd.Flags().StringP("mode", "m", "interface", "Generation mode: 'interface' or 'struct'. Ignored if --name is used.")
}
