package cli

import (
	"errors"

	"github.com/graingo/maltose/cmd/maltose/internal/gen"
	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

// daoCmd represents the dao command
var daoCmd = &cobra.Command{
	Use:   "dao",
	Short: "Generate a DAO layer from database schema",
	Long:  "Connects to a database and generates data access objects for the existing tables and entity models.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		utils.PrintInfo("✍️  Generating DAO layer...", nil)

		dst, _ := cmd.Flags().GetString("dst")

		generator, err := gen.NewDaoGenerator(dst)
		if err != nil {
			return err
		}
		if err := generator.Gen(); err != nil {
			if errors.Is(err, gen.ErrEnvFileNeedUpdate) {
				return nil
			}
			return err
		}

		utils.PrintSuccess("✅ Successfully generated DAO layer.", nil)
		return nil
	},
}

func init() {
	genCmd.AddCommand(daoCmd)

	daoCmd.Flags().StringP("dst", "d", "internal/dao", "Destination path for generated files")
}
