package gen

import (
	"fmt"
	"path/filepath"

	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
)

// ModelGenerator holds the configuration for generating model files.
type ModelGenerator struct {
}

// NewModelGenerator creates a new ModelGenerator.
func NewModelGenerator() *ModelGenerator {
	return &ModelGenerator{}
}

// Gen generates only the entity files.
func (g *ModelGenerator) Gen() error {
	if err := initDB(); err != nil {
		return err
	}

	utils.PrintInfo("entity_files_generation_start", nil)
	for _, table := range tables {
		structName := strcase.ToCamel(inflection.Singular(table.Name))
		data := daoTplData{
			TableName:  table.Name,
			StructName: structName,
			Columns:    table.Columns,
		}

		outputPath := filepath.Join("internal", "model", "entity", fmt.Sprintf("%s.go", table.Name))

		utils.PrintInfo("generating_file", utils.TplData{"Path": outputPath})
		if err := generateFile(outputPath, "entity", TplGenEntity, data); err != nil {
			return err
		}
	}
	return nil
}
