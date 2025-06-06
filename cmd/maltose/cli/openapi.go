package cli

import (
	"fmt"
	"os"

	"github.com/graingo/maltose/cmd/maltose/internal/openapi"
	"github.com/spf13/cobra"
)

var openapiCmd = &cobra.Command{
	Use:   "openapi [dir]",
	Short: "Generates OpenAPI V3 documentation from Go source files.",
	Long: `Generates OpenAPI V3 documentation by parsing Go source files in the specified directory.
It looks for structs with an embedded m.Meta field to define API endpoints.

Default directory: ./api
Default output file: ./openapi.yaml
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Determine the source directory
		srcDir := "./api"
		if len(args) > 0 {
			srcDir = args[0]
		}

		// Define the output file path
		outputFile := "./openapi.yaml"

		fmt.Printf("Generating OpenAPI spec from: %s\n", srcDir)
		if err := openapi.Generate(srcDir, outputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully generated OpenAPI specification to %s\n", outputFile)
	},
}

func init() {
	rootCmd.AddCommand(openapiCmd)
}
