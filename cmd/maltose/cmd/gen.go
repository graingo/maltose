package cmd

import (
	"fmt"
	"os"

	"github.com/graingo/maltose/cmd/maltose/internal/gendao"
	"github.com/spf13/cobra"
)

var onlyFlag string

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate model and DAO layer code from database.",
	Long: `A powerful code generation command that connects to your database to create
GORM models and a layered DAO structure.

By default, it runs the full generation process:
1. Generates models from your database schema into 'internal/model/entity'.
2. Generates a layered DAO based on those models into 'internal/dao'.

Use the --only flag to run a specific part of the process.`,
	Run: func(cmd *cobra.Command, args []string) {
		switch onlyFlag {
		case "model":
			fmt.Println("Running 'model' generation only...")
			if err := gendao.GenerateModel(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		case "dao":
			fmt.Println("Running 'dao' generation only...")
			if err := gendao.GenerateDao(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		case "":
			fmt.Println("Running full generation process (model and dao)...")
			fmt.Println("\nStep 1: Generating Models...")
			if err := gendao.GenerateModel(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✅ Models generated successfully.")

			fmt.Println("\nStep 2: Generating DAO Layer...")
			if err := gendao.GenerateDao(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✅ DAO layer generated successfully.")
		default:
			fmt.Fprintf(os.Stderr, "Error: invalid value for --only flag. Allowed values are 'model' or 'dao'.\n")
			os.Exit(1)
		}

		fmt.Println("\n✨ Code generation complete!")
	},
}

func init() {
	rootCmd.AddCommand(genCmd)
	genCmd.Flags().StringVar(&onlyFlag, "only", "", "Specify to run only 'model' or 'dao' generation.")
}
