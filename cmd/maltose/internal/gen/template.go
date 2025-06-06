package gen

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/jinzhu/inflection"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

type templateData struct {
	TableName       string
	StructName      string
	Columns         []*gorm.ColumnType
	Comment         string
	PackageName     string
	DaoName         string
	InternalDaoName string
}

// generateFile creates a file based on a template.
func generateFile(data templateData, tpl *template.Template, outputPath string) error {
	// Ensure the directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template for %s: %w", data.StructName, err)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		log.Printf("failed to format source for %s, writing unformatted code. Error: %v", outputPath, err)
		formatted = buf.Bytes() // Write unformatted code on error
	}

	// Write the file
	return os.WriteFile(outputPath, formatted, 0644)
}

// Helpers for templates
var funcMap = template.FuncMap{
	"toCamel":    toCamel,
	"toSingular": inflection.Singular,
	"dbTypeToGo": dbTypeToGo,
	"makeTags":   makeTags,
}

func toCamel(s string) string {
	caser := cases.Title(language.English)
	return caser.String(s)
}

// dbTypeToGo converts database column types to Go types.
func dbTypeToGo(column gorm.ColumnType) string {
	// This is a simplified mapping. A real implementation needs to be more robust.
	t := column.DatabaseTypeName()
	switch t {
	case "VARCHAR", "TEXT", "CHAR", "LONGTEXT":
		return "string"
	case "INT", "INTEGER", "BIGINT", "SMALLINT", "TINYINT":
		if nullable, ok := column.Nullable(); ok && nullable {
			return "*int"
		}
		return "int"
	case "TIMESTAMP", "DATETIME", "DATE":
		return "time.Time"
	case "FLOAT", "DOUBLE", "DECIMAL":
		return "float64"
	case "BOOL":
		return "bool"
	default:
		return "string"
	}
}

// makeTags creates gorm and json struct tags for a field.
func makeTags(column gorm.ColumnType) string {
	return fmt.Sprintf(`gorm:"column:%s" json:"%s"`, column.Name(), column.Name())
}

var entityTemplate = `// ==========================================================================
// Code generated and maintained by Maltose tool. DO NOT EDIT.
// ==========================================================================

package entity

import (
	"time"
)

// {{.StructName}} is the golang structure for table {{.TableName}}.
type {{.StructName}} struct {
{{- range .Columns}}
    {{toCamel .Name}} {{dbTypeToGo .}} ` + "`{{makeTags .}}`" + `
{{- end}}
}

func ({{.StructName}}) TableName() string {
    return "{{.TableName}}"
}
`

var daoInternalTemplate = `// ==========================================================================
// Code generated and maintained by Maltose tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"
	"{{.PackageName}}/internal/model/entity"
	"gorm.io/gorm"
)

type {{.InternalDaoName}} struct {
	db *gorm.DB
}

func New{{.InternalDaoName}}(db *gorm.DB) *{{.InternalDaoName}} {
	return &{{.InternalDaoName}}{db}
}

// Ctx returns a new transaction DAO.
func (dao *{{.InternalDaoName}}) Ctx(ctx context.Context) *{{.InternalDaoName}} {
	return New{{.InternalDaoName}}(dao.db.WithContext(ctx))
}

// DB returns the underlying database connection.
func (dao *{{.InternalDaoName}}) DB() *gorm.DB {
	return dao.db
}

// Create creates a new record.
func (dao *{{.InternalDaoName}}) Create(data *entity.{{.StructName}}) error {
	return dao.db.Create(data).Error
}

// GetByID retrieves a record by its primary key.
func (dao *{{.InternalDaoName}}) GetByID(id interface{}) (*entity.{{.StructName}}, error) {
	var result entity.{{.StructName}}
	err := dao.db.First(&result, id).Error
	return &result, err
}

// Update updates a record by its primary key.
// It only updates non-zero fields in the data struct.
func (dao *{{.InternalDaoName}}) Update(data *entity.{{.StructName}}) error {
	return dao.db.Model(data).Updates(data).Error
}

// Delete removes a record by its primary key.
func (dao *{{.InternalDaoName}}) Delete(id interface{}) error {
	return dao.db.Delete(&entity.{{.StructName}}{}, id).Error
}

// FindOne retrieves a single record that matches the given conditions.
// "condition" can be a struct or a map.
func (dao *{{.InternalDaoName}}) FindOne(condition ...interface{}) (*entity.{{.StructName}}, error) {
	var result entity.{{.StructName}}
	err := dao.db.Where(condition[0], condition[1:]...).First(&result).Error
	return &result, err
}

// FindAll retrieves all records that match the given conditions.
// "condition" can be a struct or a map.
func (dao *{{.InternalDaoName}}) FindAll(condition ...interface{}) ([]*entity.{{.StructName}}, error) {
	var results []*entity.{{.StructName}}
	err := dao.db.Where(condition[0], condition[1:]...).Find(&results).Error
	return results, err
}
`

var daoTemplate = `// ============================================================================
// This is auto-generated by Maltose tool but you can edit it as you like.
// ============================================================================

package dao

import (
	"{{.PackageName}}/internal/dao/internal"
	"gorm.io/gorm"
)

type {{.DaoName}} struct {
	*internal.{{.InternalDaoName}}
}

var (
	// {{.StructName}} is the DAO instance for table {{.TableName}}.
	{{.StructName}} *{{.DaoName}}
)

func New{{.DaoName}}(db *gorm.DB) {
	{{.StructName}} = &{{.DaoName}}{
		internal.New{{.InternalDaoName}}(db),
	}
}
`
