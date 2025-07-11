package cli

import (
	"path/filepath"
	"strings"

	"github.com/graingo/maltose/cmd/maltose/internal/openapi"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

var openapiCmd = &cobra.Command{
	Use:   "openapi",
	Short: "Generate OpenAPI v3 specification.",
	Long: `This command generates an OpenAPI v3 specification file by parsing Go source files.
It helps in documenting your API in a standard format.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.PrintInfo("✍️  Generating OpenAPI specification...", nil)

		src, _ := cmd.Flags().GetString("src")
		outputFile, _ := cmd.Flags().GetString("output")
		format, _ := cmd.Flags().GetString("format")

		// Auto-detect format from output file extension if not explicitly set
		if !cmd.Flags().Changed("format") {
			ext := strings.ToLower(filepath.Ext(outputFile))
			if ext == ".json" {
				format = "json"
			} else {
				format = "yaml" // Default to yaml
			}
		}

		if err := openapi.Generate(src, outputFile, format); err != nil {
			return err
		}

		utils.PrintSuccess("✅ Successfully generated OpenAPI specification to '{{.OutputFile}}'.", utils.TplData{"OutputFile": outputFile})
		return nil
	},
}

func init() {
	genCmd.AddCommand(openapiCmd)

	openapiCmd.Flags().StringP("src", "s", "api", "Source directory to parse for OpenAPI specs")
	openapiCmd.Flags().StringP("output", "o", "openapi.yaml", "Output file for OpenAPI spec")
	openapiCmd.Flags().StringP("format", "f", "yaml", "Output format: yaml or json")
}
