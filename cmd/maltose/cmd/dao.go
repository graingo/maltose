package cmd

import (
	"fmt"
	"os"

	"github.com/graingo/maltose/cmd/maltose/internal/gendao"
	"github.com/spf13/cobra"
)

// daoCmd represents the dao command
var daoCmd = &cobra.Command{
	Use:   "dao",
	Short: "Generate GORM model and DAO layer from database schema.",
	Long: `Connects to the database specified in the .env file in the project root,
and generates GORM models and a DAO layer based on the GoFrame best practices.

It creates/updates the following directories:
- internal/model/entity
- internal/dao/internal
- internal/dao
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting DAO layer generation...")

		if err := gendao.Generate(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ DAO layer generated successfully.")
	},
}

func init() {
	genCmd.AddCommand(daoCmd)
}
