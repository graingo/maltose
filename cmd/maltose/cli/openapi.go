package cli

import (
	"github.com/graingo/maltose/cmd/maltose/internal/openapi"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

var openapiCmd = &cobra.Command{
	Use:   "openapi",
	Short: utils.Print("openapi_cmd_short"),
	Long:  utils.Print("openapi_cmd_long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.PrintInfo("openapi_generation_start", nil)

		src, _ := cmd.Flags().GetString("src")
		outputFile, _ := cmd.Flags().GetString("output")

		if err := openapi.Generate(src, outputFile); err != nil {
			return err
		}

		utils.PrintSuccess("openapi_generation_success", utils.TplData{"Output": outputFile})
		return nil
	},
}

func init() {
	genCmd.AddCommand(openapiCmd)

	openapiCmd.Flags().StringP("src", "s", "api", "Source directory to parse for OpenAPI specs")
	openapiCmd.Flags().StringP("output", "o", "openapi.yaml", "Output file for OpenAPI spec")
}
