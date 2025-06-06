package cli

import (
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
		PrintInfo("Starting DAO layer generation...\n")
		if err := gen.GenerateDao(); err != nil {
			PrintError("%v\n", err)
			os.Exit(1)
		}
		PrintSuccess("✅ DAO layer generated successfully.\n")
	},
}

func init() {
	genCmd.AddCommand(daoCmd)
}
