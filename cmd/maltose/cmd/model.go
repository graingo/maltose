package cmd

import (
	"fmt"
	"os"

	"github.com/graingo/maltose/cmd/maltose/internal/gendao"
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
		fmt.Println("Starting GORM model generation...")

		// We will create a dedicated GenerateModel function later
		if err := gendao.GenerateModel(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ GORM models generated successfully in 'internal/model/entity'.")
	},
}

func init() {
	genCmd.AddCommand(modelCmd)
}
