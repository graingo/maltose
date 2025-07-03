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
	Short: "Generate DAO layer based on existing models.",
	Long:  "This command scans for GORM models and generates a complete data access object (DAO) layer, including interfaces and implementations.",
	RunE: func(cmd *cobra.Command, args []string) error {
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
