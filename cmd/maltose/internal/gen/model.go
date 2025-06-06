package gen

import (
	"fmt"
	"path/filepath"

	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
)

// GenerateModel generates only the entity files.
func GenerateModel() error {
	if err := initDB(); err != nil {
		return err
	}

	utils.PrintInfo("entityFilesGenerationStart", nil)
	for _, table := range tables {
		structName := strcase.ToCamel(inflection.Singular(table.Name))
		data := daoTplData{
			TableName:  table.Name,
			StructName: structName,
			Columns:    table.Columns,
		}

		outputPath := filepath.Join("internal", "model", "entity", fmt.Sprintf("%s.go", table.Name))

		utils.PrintInfo("generatingFile", map[string]interface{}{"Path": outputPath})
		if err := generateFile(outputPath, "entity", TplGenEntity, data); err != nil {
			return err
		}
	}
	return nil
}
