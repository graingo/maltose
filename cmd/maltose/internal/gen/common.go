// Package gen contains the common logic for code generation.
package gen

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/graingo/maltose/errors/merror"
	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

// generateFile creates a file based on a template.
func generateFile(path, tplName, tplContent string, data interface{}) error {
	// Create the template
	tpl, err := template.New(tplName).Funcs(funcMap).Parse(tplContent)
	if err != nil {
		return merror.Wrapf(err, "failed to parse template %s", tplName)
	}

	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return merror.Wrapf(err, "failed to create directory %s", dir)
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return merror.Wrapf(err, "failed to execute template %s", tplName)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		utils.PrintWarn("failed to format source for {{.Path}}, writing unformatted code. Error: {{.Error}}", utils.TplData{"Path": path, "Error": err})
		formatted = buf.Bytes() // Write unformatted code on error
	}

	// Write the file
	return os.WriteFile(path, formatted, 0644)
}

// funcMap contains helper functions for the templates.
var funcMap = template.FuncMap{
	"toCamel":     strcase.ToCamel,
	"toSingular":  inflection.Singular,
	"dbTypeToGo":  dbTypeToGo,
	"makeTags":    makeTags,
	"makeRemarks": makeRemarks,
	"firstLower":  strcase.ToLowerCamel,
	"toTitle":     cases.Title(language.English).String,
}

// dbTypeToGo converts database column types to Go types.
func dbTypeToGo(column gorm.ColumnType) string {
	t := strings.ToUpper(column.DatabaseTypeName())

	if strings.Contains(t, "UNSIGNED") {
		switch {
		case strings.HasPrefix(t, "TINYINT"):
			return "uint8"
		case strings.HasPrefix(t, "SMALLINT"):
			return "uint16"
		case strings.HasPrefix(t, "MEDIUMINT"):
			return "uint32"
		case strings.HasPrefix(t, "INT"):
			return "uint"
		case strings.HasPrefix(t, "BIGINT"):
			return "uint64"
		}
	}

	switch {
	case strings.HasPrefix(t, "INT"):
		return "int"
	case strings.HasPrefix(t, "TINYINT"):
		return "int8"
	case strings.HasPrefix(t, "SMALLINT"):
		return "int16"
	case strings.HasPrefix(t, "BIGINT"):
		return "int64"
	}

	switch t {
	case "VARCHAR", "TEXT", "CHAR", "LONGTEXT", "JSON":
		return "string"
	case "TIMESTAMP", "DATETIME", "DATE", "TIME":
		return "time.Time"
	case "FLOAT", "DOUBLE":
		return "float64"
	case "DECIMAL", "NUMERIC":
		return "decimal.Decimal"
	case "BOOL", "BOOLEAN":
		return "bool"
	case "BLOB", "LONGBLOB", "BINARY", "VARBINARY":
		return "[]byte"
	default:
		return "string"
	}
}

// makeTags creates gorm and json struct tags for a field.
func makeTags(column gorm.ColumnType) string {
	return fmt.Sprintf(`gorm:"column:%s" json:"%s"`, column.Name(), strcase.ToLowerCamel(column.Name()))
}

// makeRemarks creates remarks for a field.
func makeRemarks(column gorm.ColumnType) string {
	comment, ok := column.Comment()
	if !ok || comment == "" {
		return ""
	}
	return fmt.Sprintf("// %s", comment)
}
