package gen

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
)

// ModelGenerator holds the configuration for generating model files.
type ModelGenerator struct {
	Dst     string
	Table   string
	Exclude string
}

// NewModelGenerator creates a new ModelGenerator.
func NewModelGenerator(dst, table, exclude string) *ModelGenerator {
	return &ModelGenerator{
		Dst:     dst,
		Table:   table,
		Exclude: exclude,
	}
}

// Gen generates only the entity files.
func (g *ModelGenerator) Gen() error {
	if err := initDB(); err != nil {
		return err
	}

	// filter tables
	var filteredTables []TableInfo
	if g.Table != "" {
		tableNames := strings.Split(g.Table, ",")
		tableSet := make(map[string]struct{})
		for _, name := range tableNames {
			tableSet[strings.TrimSpace(name)] = struct{}{}
		}
		for _, table := range tables {
			if _, ok := tableSet[table.Name]; ok {
				filteredTables = append(filteredTables, table)
			}
		}
	} else if g.Exclude != "" {
		excludeNames := strings.Split(g.Exclude, ",")
		excludeSet := make(map[string]struct{})
		for _, name := range excludeNames {
			excludeSet[strings.TrimSpace(name)] = struct{}{}
		}
		for _, table := range tables {
			if _, ok := excludeSet[table.Name]; !ok {
				filteredTables = append(filteredTables, table)
			}
		}
	} else {
		filteredTables = tables
	}

	utils.PrintInfo("entity_files_generation_start", nil)
	for _, table := range filteredTables {
		structName := strcase.ToCamel(inflection.Singular(table.Name))

		hasTime := false
		hasDecimal := false
		for _, column := range table.Columns {
			goType := dbTypeToGo(column)
			if goType == "time.Time" {
				hasTime = true
			}
			if goType == "decimal.Decimal" {
				hasDecimal = true
			}
		}

		data := daoTplData{
			TableName:  table.Name,
			StructName: structName,
			Columns:    table.Columns,
			HasTime:    hasTime,
			HasDecimal: hasDecimal,
		}

		outputPath := filepath.Join(g.Dst, "entity", fmt.Sprintf("%s.go", table.Name))

		utils.PrintInfo("generating_file", utils.TplData{"Path": outputPath})
		if err := generateFile(outputPath, "entity", TplGenEntity, data); err != nil {
			return err
		}
	}
	return nil
}
