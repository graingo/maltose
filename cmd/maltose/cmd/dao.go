package cmd

import (
	"fmt"
	"os"

	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/spf13/cobra"
)

// daoCmd represents the dao command
var daoCmd = &cobra.Command{
	Use:   "dao",
	Short: "Generate DAO layer based on existing models.",
	Long: `Scans for models in 'internal/model/entity' and generates a complete
DAO layer in 'internal/dao' and 'internal/dao/internal'.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting DAO layer generation...")
		if err := gen.GenerateDao(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ DAO layer generated successfully.")
	},
}

func init() {
	genCmd.AddCommand(daoCmd)
}
