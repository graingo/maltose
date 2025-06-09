package cli

import (
	"os"

	"github.com/graingo/maltose/cmd/maltose/internal/openapi"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

var (
	openapiOutputFile string
	openapiSrcPath    string
)

var openapiCmd = &cobra.Command{
	Use:   "openapi",
	Short: utils.Print("openapi_cmd_short"),
	Long:  utils.Print("openapi_cmd_long"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := openapi.Generate(openapiSrcPath, openapiOutputFile); err != nil {
			utils.PrintError("generic_error", utils.TplData{"Error": err})
			os.Exit(1)
		}

		utils.PrintSuccess("openapi_generation_success", utils.TplData{"Output": openapiOutputFile})
	},
}

func init() {
	rootCmd.AddCommand(openapiCmd)

	openapiCmd.Flags().StringVarP(&openapiSrcPath, "src", "s", "api", "Source directory to parse for OpenAPI specs")
	openapiCmd.Flags().StringVarP(&openapiOutputFile, "output", "o", "openapi.yaml", "Output file for OpenAPI spec")
}
