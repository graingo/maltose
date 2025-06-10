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
	Dst        string
	ModulePath string
	ModuleRoot string
}

// NewDaoGenerator creates a new DaoGenerator.
func NewDaoGenerator(dst string) (*DaoGenerator, error) {
	moduleName, moduleRoot, err := utils.GetModuleInfo(dst)
	if err != nil {
		return nil, fmt.Errorf("failed to get module info: %w", err)
	}

	return &DaoGenerator{
		Dst:        dst,
		ModulePath: moduleName,
		ModuleRoot: moduleRoot,
	}, nil
}

// daoTplData holds all the template variables for generating DAO and entity files.
type daoTplData struct {
	TableName       string
	StructName      string
	PackageName     string
	InternalDaoName string
	DaoName         string
	Columns         []gorm.ColumnType
	HasTime         bool
	HasDecimal      bool
}

// Gen generates only the DAO files.
func (g *DaoGenerator) Gen() error {
	if err := initDB(); err != nil {
		return err
	}

	utils.PrintInfo("dao_files_generation_start", nil)

	for _, table := range tables {
		structName := strcase.ToCamel(inflection.Singular(table.Name))
		daoName := structName + "Dao"
		data := daoTplData{
			TableName:   table.Name,
			StructName:  structName,
			PackageName: g.ModulePath,
			DaoName:     daoName,
		}

		internalPath := filepath.Join(g.Dst, "internal", fmt.Sprintf("%s.go", table.Name))
		utils.PrintInfo("generating_file", utils.TplData{"Path": internalPath})
		if err := generateFile(internalPath, "daoInternal", TplGenDaoInternal, data); err != nil {
			return err
		}

		daoPath := filepath.Join(g.Dst, fmt.Sprintf("%s.go", table.Name))

		if _, err := os.Stat(daoPath); os.IsNotExist(err) {
			utils.PrintInfo("generating_file", utils.TplData{"Path": daoPath})
			if err := generateFile(daoPath, "dao", TplGenDao, data); err != nil {
				return err
			}
		} else {
			utils.PrintWarn("skipping_file", utils.TplData{"Path": daoPath})
		}
	}
	return nil
}
