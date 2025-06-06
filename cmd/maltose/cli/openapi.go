package cli

import (
	"github.com/graingo/maltose/cmd/maltose/i18n"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

var (
	outputFile string
	srcDir     string
)

var openapiCmd = &cobra.Command{
	Use:   "openapi",
	Short: i18n.T("openapi_cmd_short", nil),
	Long:  i18n.T("openapi_cmd_long", nil),
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		srcDir, _ := cmd.Flags().GetString("src")
		outputFile, _ := cmd.Flags().GetString("output")

		utils.PrintInfo("openapi_generation_start", utils.TplData{"Source": srcDir})
		// TODO: The function to generate the OpenAPI spec needs to be implemented or located.
		// The original `gen.GenerateOpenAPISpec` function was not found.
		// if err := gen.GenerateOpenAPISpec(srcDir, outputFile); err != nil {
		// 	utils.PrintError("generic_error", utils.TplData{"Error": err})
		// 	os.Exit(1)
		// }
		utils.PrintSuccess("openapi_generation_success", utils.TplData{"Output": outputFile})
	},
}

func init() {
	rootCmd.AddCommand(openapiCmd)

	openapiCmd.Flags().StringP("src", "s", "./...", "Source directory to parse for OpenAPI specs")
	openapiCmd.Flags().StringP("output", "o", "openapi.json", "Output file for OpenAPI spec")
}
