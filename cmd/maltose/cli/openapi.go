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
	Short: "Generate an OpenAPI document and contract manifest.",
	Long: `Generate OpenAPI and its contract manifest using the application's Maltose compiler.
API packages must compile and should keep init functions free of side effects.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
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

		version, _ := cmd.Flags().GetString("openapi-version")
		extensions, _ := cmd.Flags().GetString("extensions")
		check, _ := cmd.Flags().GetBool("check")
		if err := openapi.Run(cmd.Context(), openapi.Config{Source: src, Output: outputFile, Format: format, Version: version, Extensions: extensions, Check: check}); err != nil {
			return err
		}

		if check {
			utils.PrintSuccess("✅ OpenAPI document and manifest are current.", nil)
			return nil
		}
		utils.PrintSuccess("✅ Successfully generated OpenAPI specification to '{{.OutputFile}}'.", utils.TplData{"OutputFile": outputFile})
		return nil
	},
}

func init() {
	genCmd.AddCommand(openapiCmd)
	openapiCmd.Flags().String("openapi-version", "3.1.0", "OpenAPI version: 3.0.0 or 3.1.0")
	openapiCmd.Flags().String("extensions", "", "Go package exporting Configure(*contract.Extensions) error")
	openapiCmd.Flags().Bool("check", false, "Check generated document and manifest without writing")

	openapiCmd.Flags().StringP("src", "s", "api", "Source directory to parse for OpenAPI specs")
	openapiCmd.Flags().StringP("output", "o", "openapi.yaml", "Output file for OpenAPI spec")
	openapiCmd.Flags().StringP("format", "f", "yaml", "Output format: yaml or json")
}
