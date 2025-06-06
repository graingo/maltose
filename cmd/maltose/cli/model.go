package cli

import (
	"os"

	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/spf13/cobra"
)

// modelCmd represents the model command
var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Generate GORM model from database schema.",
	Long: `Connects to the database specified in the .env file in the project root
and generates GORM models in 'internal/model/entity'.
`,
	Run: func(cmd *cobra.Command, args []string) {
		PrintInfo("Starting GORM model generation...\n")
		if err := gen.GenerateModel(); err != nil {
			PrintError("%v\n", err)
			os.Exit(1)
		}
		PrintSuccess("✅ GORM models generated successfully.\n")
	},
}

func init() {
	genCmd.AddCommand(modelCmd)
}
