package gen

const (
	ControllerTemplate = `package controller

import (
	"{{.Module}}/internal/service"
	"context"
)

type {{.StructName}}Controller struct{}

func (c *{{.StructName}}Controller) Get(ctx context.Context, req *v1.{{.StructName}}GetReq) (res *v1.{{.StructName}}GetRes, err error) {
	// TODO: Add your logic here
	return
}
`

	ServiceInterfaceTemplate = `package service

import (
	"context"
)

type I{{.StructName}} interface {
    Get(ctx context.Context, req *v1.{{.StructName}}GetReq) (res *v1.{{.StructName}}GetRes, err error)
}

var local{{.StructName}} I{{.StructName}}

func {{.StructName}}() I{{.StructName}} {
	if local{{.StructName}} == nil {
		panic("implement not found for interface I{{.StructName}}, forgot register?")
	}
	return local{{.StructName}}
}

func Register{{.StructName}}(i I{{.StructName}}) {
	local{{.StructName}} = i
}
`

	ServiceStructTemplate = `package service

import (
	"context"
)

type {{.StructName}}Service struct{}

func New{{.StructName}}Service() *{{.StructName}}Service {
	return &{{.StructName}}Service{}
}

func (s *{{.StructName}}Service) Get(ctx context.Context, req *v1.{{.StructName}}GetReq) (res *v1.{{.StructName}}GetRes, err error) {
	// TODO: Add your logic here
	return
}
`
)
