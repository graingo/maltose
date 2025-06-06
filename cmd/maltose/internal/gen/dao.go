package gen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"gorm.io/gorm"
)

// daoTplData holds all the template variables for generating DAO and entity files.
type daoTplData struct {
	TableName       string
	StructName      string
	PackageName     string
	InternalDaoName string
	DaoName         string
	Columns         []gorm.ColumnType
}

// GenerateDao generates only the DAO files.
func GenerateDao() error {
	if err := initDB(); err != nil {
		return err
	}

	utils.PrintInfo("daoFilesGenerationStart", nil)
	modulePath, err := getGoModulePath()
	if err != nil {
		return fmt.Errorf("failed to get go module path: %w", err)
	}

	for _, table := range tables {
		structName := strcase.ToCamel(inflection.Singular(table.Name))
		daoName := structName + "Dao"
		data := daoTplData{
			TableName:   table.Name,
			StructName:  structName,
			PackageName: modulePath,
			DaoName:     daoName,
		}

		internalPath := filepath.Join("internal", "dao", "internal", fmt.Sprintf("%s.go", table.Name))
		utils.PrintInfo("generatingFile", map[string]interface{}{"Path": internalPath})
		if err := generateFile(internalPath, "daoInternal", TplGenDaoInternal, data); err != nil {
			return err
		}

		daoPath := filepath.Join("internal", "dao", fmt.Sprintf("%s.go", table.Name))
		if _, err := os.Stat(daoPath); os.IsNotExist(err) {
			utils.PrintInfo("generatingFile", map[string]interface{}{"Path": daoPath})
			if err := generateFile(daoPath, "dao", TplGenDao, data); err != nil {
				return err
			}
		} else {
			utils.PrintInfo("skippingFile", map[string]interface{}{"Path": daoPath})
		}
	}
	return nil
}

func getGoModulePath() (string, error) {
	cmd := exec.Command("go", "list", "-m")
	var out strings.Builder
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}
