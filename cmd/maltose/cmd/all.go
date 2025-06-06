package cmd

import (
	"fmt"
	"os"

	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/spf13/cobra"
)

// allCmd represents the all command
var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Generate both model and DAO layers.",
	Long:  `A convenient command that runs the full generation process, generating both models and the DAO layer.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running full generation process (model and dao)...")

		fmt.Println("\nStep 1: Generating Models...")
		if err := gen.GenerateModel(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Models generated successfully.")

		fmt.Println("\nStep 2: Generating DAO Layer...")
		if err := gen.GenerateDao(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ DAO layer generated successfully.")

		fmt.Println("\n✨ Code generation complete!")
	},
}

func init() {
	genCmd.AddCommand(allCmd)
}
