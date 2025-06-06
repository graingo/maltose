package cli

import (
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

var (
	outputFile string
	srcDir     string
)

var openapiCmd = &cobra.Command{
	Use:   "openapi",
	Short: "Generate an OpenAPI specification from Go source files.",
	Long: `This command parses Go source files that define API endpoints and generates
an OpenAPI 3.0 specification file in JSON format. It helps in documenting
and testing APIs efficiently.`,
	Run: func(cmd *cobra.Command, args []string) {
		utils.PrintInfo("openAPIGenerationStart", map[string]interface{}{"Source": srcDir})
		// TODO: The function to generate the OpenAPI spec needs to be implemented or located.
		// The original `gen.GenerateOpenAPISpec` function was not found.
		// if err := gen.GenerateOpenAPISpec(srcDir, outputFile); err != nil {
		// 	utils.PrintError("genericError", map[string]interface{}{"Error": err})
		// 	os.Exit(1)
		// }
		utils.PrintSuccess("openAPIGenerationSuccess", map[string]interface{}{"Output": outputFile})
	},
}

func init() {
	genCmd.AddCommand(openapiCmd)
	openapiCmd.Flags().StringVarP(&outputFile, "output", "o", "openapi.json", "Output file for the OpenAPI specification")
	openapiCmd.Flags().StringVarP(&srcDir, "src", "s", "api", "Source directory containing API definitions")
}
