// Package gen provides the generation of go code.
package gen

const (
	// TplGenController is the template for generating controller files.
	// This is for the simple case: api/<version>/<file>.go
	TplGenController = `// =================================================================================
	// Code generated and maintained by Maltose tool. You can edit this file as you like.
	// =================================================================================
	package v1

	import (
		"context"
		"{{.ApiModule}}"
	)

	type c{{.Service}} struct{}

	// New{{.Service}} creates a new controller.
	func New{{.Service}}() *c{{.Service}} {
		return &c{{.Service}}{}
	}

	{{range .Functions}}
	// {{.Name}} is the handler for the {{.Name}} API.
	func (c *c{{$.Service}}) {{.Name}}(ctx context.Context, req *{{$.ApiPkg}}.{{.ReqName}}) (res *{{$.ApiPkg}}.{{.ResName}}, err error) {
		// TODO: Implement the business logic here.
		panic("implement me")
	}
	{{end}}
`

	// TplGenControllerStruct is the template for the controller struct definition file.
	// Used for the case: api/<module>/<version>/...
	TplGenControllerStruct = `// =================================================================================
	// Code generated and maintained by Maltose tool. You can edit this file as you like.
	// =================================================================================
	package {{.Module}}

	type {{.Controller}} struct{}

	func New{{.Version}}() *{{.Controller}} {
		return &{{.Controller}}{}
	}
`

	// TplGenControllerMethod is the template for the controller method implementation file.
	// Used for the case: api/<module>/<version>/...
	TplGenControllerMethod = `// =================================================================================
	// Code generated and maintained by Maltose tool. You can edit this file as you like.
	// =================================================================================
	package {{.Module}}

	import (
		"context"
		"{{.ApiModule}}"
	)

	{{range .Functions}}
	// {{.Name}} is the handler for the {{.Name}} API.
	func (c *{{$.Controller}}) {{.Name}}(ctx context.Context, req *{{$.ApiPkg}}.{{.ReqName}}) (res *{{$.ApiPkg}}.{{.ResName}}, err error) {
		// TODO: Implement the business logic here.
		panic("implement me")
	}
	{{end}}
`

	// TplGenControllerMethodOnly is the template for appending new methods to an existing controller file.
	TplGenControllerMethodOnly = `
{{range .Functions}}
// {{.Name}} is the handler for the {{.Name}} API.
func (c *{{$.Controller}}) {{.Name}}(ctx context.Context, req *{{$.ApiPkg}}.{{.ReqName}}) (res *{{$.ApiPkg}}.{{.ResName}}, err error) {
	// TODO: Implement the business logic here.
	panic("implement me")
}
{{end}}
`

	// TplGenService is the template for generating service files.
	TplGenService = `// =================================================================================
	// Code generated and maintained by Maltose tool. You can edit this file as you like.
	// =================================================================================
	package service

	type s{{.Service}} struct{}

	var local{{.Service}} = New{{.Service}}()

	// New{{.Service}} creates a new service instance.
	func New{{.Service}}() *s{{.Service}} {
		return &s{{.Service}}{}
	}

	// {{.Service}} returns the default service instance.
	func {{.Service}}() *s{{.Service}} {
		return local{{.Service}}
	}
`

	// TplGenServiceMethodOnly is the template for appending new methods to an existing service file.
	TplGenServiceMethodOnly = `
{{range .Functions}}
// {{.Name}} is the handler for the {{.Name}} API.
func (s *s{{$.Service}}) {{.Name}}(ctx context.Context, req *{{$.ApiPkg}}.{{.ReqName}}) (res *{{$.ApiPkg}}.{{.ResName}}, err error) {
	// TODO: Implement the business logic here.
	panic("implement me")
}
{{end}}
`

	// TplGenServiceInterface is the template for the service interface.
	TplGenServiceInterface = `// =================================================================================
	// Code generated and maintained by Maltose tool. You can edit this file as you like.
	// =================================================================================
	package service

	type I{{.Service}} interface {
		// TODO: Define your service interface methods here.
	}

	var local{{.Service}} I{{.Service}}

	// {{.Service}} returns the registered implementation of I{{.Service}}.
	// It panics if no implementation is registered.
	func {{.Service}}() I{{.Service}} {
		if local{{.Service}} == nil {
			panic("implement not found for interface I{{.Service}}, forgot register?")
		}
		return local{{.Service}}
	}

	// Register{{.Service}} registers an implementation for the I{{.Service}} interface.
	func Register{{.Service}}(i I{{.Service}}) {
		local{{.Service}} = i
	}
`

	// TplGenServiceInterfaceMethodOnly is the template for appending new methods to an existing service interface file.
	TplGenServiceInterfaceMethodOnly = `
{{range .Functions}}
	{{.Name}}(ctx context.Context, req *{{$.ApiPkg}}.{{.ReqName}}) (res *{{$.ApiPkg}}.{{.ResName}}, err error)
{{end}}
`

	// TplGenServiceLogic is the template for the service logic implementation.
	TplGenServiceLogic = `// =================================================================================
	// Code generated and maintained by Maltose tool. You can edit this file as you like.
	// =================================================================================
	package {{.Module}}

	import (
		"context"

		"{{.ApiModule}}"
		"{{.SvcPackage}}"
	)

	func init() {
		service.Register{{.Service}}(New())
	}

	type s{{.Service}} struct{}

	// New creates a new service logic implementation.
	func New() service.I{{.Service}} {
		return &s{{.Service}}{}
	}

	{{range .Functions}}
	func (s *s{{$.Service}}) {{.Name}}(ctx context.Context{{if .ReqName}}, input {{if .ReqIsPointer}}*{{end}}{{$.ApiPkg}}.{{.ReqName}}{{end}}) ({{if .ResName}}output {{if .ResIsPointer}}*{{end}}{{$.ApiPkg}}.{{.ResName}}, {{end}}err error) {
		// TODO: Implement the business logic of {{.Name}}.
		{{if .ResName}}{{if .ResIsPointer}}output = new({{$.ApiPkg}}.{{.ResName}}){{end}}{{end}}
		return
	}
	{{end}}
`

	// TplGenServiceLogicAppend is the template for appending new methods to a service logic file.
	TplGenServiceLogicAppend = `
{{range .Functions}}
func (s *s{{$.Service}}) {{.Name}}(ctx context.Context{{if .ReqName}}, input {{if .ReqIsPointer}}*{{end}}{{$.ApiPkg}}.{{.ReqName}}{{end}}) ({{if .ResName}}output {{if .ResIsPointer}}*{{end}}{{$.ApiPkg}}.{{.ResName}}, {{end}}err error) {
	// TODO: Implement the business logic of {{.Name}}.
	{{if .ResName}}{{if .ResIsPointer}}output = new({{$.ApiPkg}}.{{.ResName}}){{end}}{{end}}
	return
}
{{end}}
`

	// TplGenEntity is the template for generating model entity files.
	TplGenEntity = `// =================================================================================
	// Code generated and maintained by Maltose tool. DO NOT EDIT.
	// =================================================================================
package entity
{{if .HasTime}}
import "time"
{{end}}
{{if .HasDecimal}}
import "github.com/shopspring/decimal"
{{end}}

// {{.StructName}} is the golang structure for table {{.TableName}}.
type {{.StructName}} struct {
{{- range .Columns}}
    {{toCamel .Name}} {{dbTypeToGo .}} ` + "`{{makeTags .}}`" + ` {{makeRemarks .}}
{{- end}}
}

	// TableName returns the name of the table.
	func (*{{.StructName}}) TableName() string {
    return "{{.TableName}}"
}
`

	// TplGenDaoInternal is the template for generating internal DAO files.
	TplGenDaoInternal = `// =================================================================================
	// Code generated and maintained by Maltose tool. DO NOT EDIT.
	// =================================================================================
package internal

import (
	"context"
		"errors"

	"github.com/graingo/maltose/database/mdb"
	"gorm.io/gorm"
	"{{.PackageName}}/internal/model/entity"
	)

	type {{.DaoName}} struct {
		DB *mdb.DB
	}

	func New{{.DaoName}}(db *mdb.DB) *{{.DaoName}} {
		return &{{.DaoName}}{DB: db}
	}

	func (d *{{.DaoName}}) Create(ctx context.Context, data *entity.{{.StructName}}) error {
		return d.DB.WithContext(ctx).Create(data).Error
	}

	// FirstOrCreate finds the first record that matches the given conditions, or creates a new one if not found.
	// The found/created record is returned.
	func (d *{{.DaoName}}) FirstOrCreate(ctx context.Context, condition map[string]any) (*entity.{{.StructName}}, error) {
		var result entity.{{.StructName}}
		err := d.DB.WithContext(ctx).Where(condition).FirstOrCreate(&result).Error
		if err != nil {
			return nil, err
		}
		return &result, nil
	}

	// Update updates a full record by its primary key.
	// It will update all fields, including zero values.
	func (d *{{.DaoName}}) Update(ctx context.Context, data *entity.{{.StructName}}) error {
		return d.DB.WithContext(ctx).Save(data).Error
	}

	// UpdateColumns updates specific columns of a record by its primary key.
	func (d *{{.DaoName}}) UpdateColumns(ctx context.Context, id any, updates map[string]any) error {
		return d.DB.WithContext(ctx).Model(&entity.{{.StructName}}{}).Where("id = ?", id).Updates(updates).Error
	}

	func (d *{{.DaoName}}) Delete(ctx context.Context, id any) error {
		return d.DB.WithContext(ctx).Delete(&entity.{{.StructName}}{}, id).Error
	}

	func (d *{{.DaoName}}) GetByID(ctx context.Context, id any) (*entity.{{.StructName}}, error) {
		var result entity.{{.StructName}}
		err := d.DB.WithContext(ctx).First(&result, id).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil // Record not found is not a system error
			}
			return nil, err
		}
		return &result, nil
}

// FindOne retrieves a single record that matches the given conditions.
	func (d *{{.DaoName}}) FindOne(ctx context.Context, condition map[string]any) (*entity.{{.StructName}}, error) {
	var result entity.{{.StructName}}
		err := d.DB.WithContext(ctx).Where(condition).First(&result).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil // Record not found is not a system error
			}
			return nil, err
		}
		return &result, nil
	}

	// FindList retrieves a list of records based on conditions, with ordering.
	func (d *{{.DaoName}}) FindList(ctx context.Context, condition map[string]any, orderBy ...string) ([]*entity.{{.StructName}}, error) {
		var list  []*entity.{{.StructName}}
		
		db := d.DB.WithContext(ctx).Model(&entity.{{.StructName}}{}).Where(condition)

		// Apply ordering and pagination
		if len(orderBy) > 0 {
			db = db.Order(orderBy[0])
		}

		// Execute the query
		err := db.Find(&list).Error
		if err != nil {
			return nil, err
		}

		return list, nil
	}

	// FindPageList retrieves a list of records based on conditions, with pagination and ordering.
	func (d *{{.DaoName}}) FindPageList(ctx context.Context, condition map[string]any, page, pageSize int, orderBy ...string) ([]*entity.{{.StructName}}, int64, error) {
		var (
			list  []*entity.{{.StructName}}
			total int64
		)
		
		db := d.DB.WithContext(ctx).Model(&entity.{{.StructName}}{}).Where(condition)

		// Get total count for pagination
		err := db.Count(&total).Error
		if err != nil {
			return nil, 0, err
		}

		// Apply ordering and pagination
		if len(orderBy) > 0 {
			db = db.Order(orderBy[0])
		}
		if page > 0 && pageSize > 0 {
			db = db.Offset((page - 1) * pageSize).Limit(pageSize)
		}

		// Execute the query
		err = db.Find(&list).Error
		if err != nil {
			return nil, 0, err
		}

		return list, total, nil
	}
	`

	// TplGenDao is the template for generating user-extendable DAO files.
	TplGenDao = `// =================================================================================
	// Code generated and maintained by Maltose tool. You can edit this file as you like.
	// =================================================================================
package dao

import (
	"github.com/graingo/maltose/database/mdb"
	"{{.PackageName}}/internal/dao/internal"
)

type {{.DaoName}} struct {
		*internal.{{.DaoName}}
}

	func New{{.DaoName}}(db *mdb.DB) *{{.DaoName}} {
		return &{{.DaoName}}{
			internal.New{{.DaoName}}(db),
		}
	}
	`

	// TplGenLogicManifest is the template for the main logic file that imports all logic packages.
	TplGenLogicManifest = `// =================================================================================
// Code generated and maintained by Maltose tool. DO NOT EDIT.
// =================================================================================
package logic

import (
	{{- range .Packages }}
	_ "{{.}}"
	{{- end }}
)
`
)
