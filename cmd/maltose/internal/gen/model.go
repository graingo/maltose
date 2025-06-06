package gen

import (
	"fmt"
	"path/filepath"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
)

// GenerateModel generates only the entity files.
func GenerateModel() error {
	if err := initDB(); err != nil {
		return err
	}

	fmt.Println("H Generating entity files...")
	for _, table := range tables {
		structName := strcase.ToCamel(inflection.Singular(table.Name))
		data := DaoTplData{
			TableName:  table.Name,
			StructName: structName,
			Columns:    table.Columns,
		}

		outputPath := filepath.Join("internal", "model", "entity", fmt.Sprintf("%s.go", table.Name))

		fmt.Printf("  -> Generating %s\n", outputPath)
		if err := generateFile(outputPath, "entity", TplGenEntity, data); err != nil {
			return err
		}
	}
	return nil
}
