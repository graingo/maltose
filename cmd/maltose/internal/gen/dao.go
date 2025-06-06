package gen

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"gorm.io/gorm"
)

// DaoGenerator holds the configuration for generating DAO files.
type DaoGenerator struct {
	ModulePath string
}

// NewDaoGenerator creates a new DaoGenerator.
func NewDaoGenerator(modulePath string) *DaoGenerator {
	return &DaoGenerator{
		ModulePath: modulePath,
	}
}

// daoTplData holds all the template variables for generating DAO and entity files.
type daoTplData struct {
	TableName       string
	StructName      string
	PackageName     string
	InternalDaoName string
	DaoName         string
	Columns         []gorm.ColumnType
}

// Gen generates only the DAO files.
func (g *DaoGenerator) Gen() error {
	if err := initDB(); err != nil {
		return err
	}

	utils.PrintInfo("daoFilesGenerationStart", nil)

	for _, table := range tables {
		structName := strcase.ToCamel(inflection.Singular(table.Name))
		daoName := structName + "Dao"
		data := daoTplData{
			TableName:   table.Name,
			StructName:  structName,
			PackageName: g.ModulePath,
			DaoName:     daoName,
		}

		internalPath := filepath.Join("internal", "dao", "internal", fmt.Sprintf("%s.go", table.Name))
		utils.PrintInfo("generatingFile", utils.TplData{"Path": internalPath})
		if err := generateFile(internalPath, "daoInternal", TplGenDaoInternal, data); err != nil {
			return err
		}

		daoPath := filepath.Join("internal", "dao", fmt.Sprintf("%s.go", table.Name))
		if _, err := os.Stat(daoPath); os.IsNotExist(err) {
			utils.PrintInfo("generatingFile", utils.TplData{"Path": daoPath})
			if err := generateFile(daoPath, "dao", TplGenDao, data); err != nil {
				return err
			}
		} else {
			utils.PrintInfo("skippingFile", utils.TplData{"Path": daoPath})
		}
	}
	return nil
}
