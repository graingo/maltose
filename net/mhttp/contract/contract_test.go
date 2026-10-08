package contract_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	first "github.com/graingo/maltose/net/mhttp/contract/testdata/first"
	second "github.com/graingo/maltose/net/mhttp/contract/testdata/second"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/graingo/maltose/net/mhttp/contract"
	"github.com/graingo/maltose/util/mmeta"
	"github.com/stretchr/testify/require"
)

type Child struct {
	Name string `json:"name" field:"required" binding:"min=2"`
}
type Common struct {
	Code string `json:"code" binding:"required,min=2"`
}
type CreateReq struct {
	Ratio      float64 `json:"ratio" binding:"oneof=1.0 2.0"`
	mmeta.Meta `method:"POST" group:"/api/v1" path:"/products/:id" status:"201" operation_id:"createProduct"`
	Common     `field:"inline"`
	ID         int              `path:"id"`
	Page       int              `query:"page" default:"1" binding:"min=1,max=100"`
	Token      string           `header:"Idempotency-Key" field:"required"`
	Count      int              `json:"count" field:"required" binding:"min=0"`
	Enabled    bool             `json:"enabled" field:"required"`
	Remark     string           `json:"remark,omitempty" binding:"min=2"`
	Mode       string           `json:"mode" binding:"oneof=task module mixed"`
	Profile    *Child           `json:"profile"`
	Optional   *Child           `json:"optional" field:"nullable"`
	Children   []Child          `json:"children"`
	Values     []int            `json:"values" binding:"dive,min=1"`
	Attributes map[string]Child `json:"attributes"`
}
type CreateRes struct {
	Level    int64 `json:"level" schema:"enum=1|2"`
	Common   `field:"inline"`
	Time     time.Time `json:"time"`
	Profile  *Child    `json:"profile"`
	Items    []Child   `json:"items"`
	Optional *string   `json:"optional,omitempty"`
}

func operation(t *testing.T) *contract.Operation {
	t.Helper()
	op, err := contract.Compile(contract.TypeOf[CreateReq](), contract.TypeOf[CreateRes]())
	require.NoError(t, err)
	return op
}
func bind(op *contract.Operation, body, query string) (CreateReq, contract.Input, error) {
	req := httptest.NewRequest("POST", "/api/v1/products/7"+query, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "key")
	var target CreateReq
	input, err := op.Bind(req, map[string]string{"id": "7"}, &target)
	return target, input, err
}
func TestPresenceAndValueRules(t *testing.T) {
	op := operation(t)
	for _, tc := range []struct{ name, body, query, field, code string }{
		{"numeric enum", `{"code":"ok","count":0,"enabled":false,"ratio":1}`, "", "", ""},
		{"zero values", `{"code":"ok","count":0,"enabled":false}`, "", "", ""},
		{"missing", `{"code":"ok","enabled":false}`, "", "json.count", "required"},
		{"zero bypass removed", `{"code":"ok","count":0,"enabled":false,"remark":""}`, "", "json.remark", "min"},
		{"non nullable", `{"code":"ok","count":0,"enabled":false,"profile":null}`, "", "json.profile", "nullable"},
		{"nullable", `{"code":"ok","count":0,"enabled":false,"optional":null}`, "", "", ""},
		{"child missing", `{"code":"ok","count":0,"enabled":false,"profile":{}}`, "", "json.profile.name", "required"},
		{"array child", `{"code":"ok","count":0,"enabled":false,"children":[{}]}`, "", "json.children[0].name", "required"},
		{"map child", `{"code":"ok","count":0,"enabled":false,"attributes":{"x":{}}}`, "", `json.attributes["x"].name`, "required"},
		{"dive", `{"code":"ok","count":0,"enabled":false,"values":[0]}`, "", "json.values[0]", "min"},
		{"unknown", `{"code":"ok","count":0,"enabled":false,"extra":1}`, "", "json.extra", "unknown"},
		{"invalid enum", `{"code":"ok","count":0,"enabled":false,"mode":"other"}`, "", "json.mode", "oneof"},
		{"explicit page zero", `{"code":"ok","count":0,"enabled":false}`, "?page=0", "query.page", "min"},
		{"invalid page", `{"code":"ok","count":0,"enabled":false}`, "?page=no", "query.page", "type"},
		{"duplicate scalar", `{"code":"ok","count":0,"enabled":false}`, "?page=1&page=2", "query.page", "type"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value, input, err := bind(op, tc.body, tc.query)
			if tc.code == "" {
				require.NoError(t, err)
				require.Equal(t, 1, value.Page)
				require.Equal(t, 7, value.ID)
				require.Equal(t, contract.Missing, input.Presence("query.page"))
				require.Equal(t, contract.Present, input.Presence("json.enabled"))
				if strings.Contains(tc.body, "null") {
					require.Equal(t, contract.Null, input.Presence("json.optional"))
				}
				return
			}
			var validation *contract.ValidationError
			require.True(t, errors.As(err, &validation), "%v", err)
			require.Equal(t, tc.field, validation.Issues[0].Field)
			require.Equal(t, tc.code, validation.Issues[0].Code)
		})
	}
	for _, body := range []string{`null`, `[]`, `{} {}`, `{"broken"`} {
		_, _, err := bind(op, body, "")
		var decode *contract.DecodeError
		require.ErrorAs(t, err, &decode)
	}
}

type Problem struct {
	Params      map[string]any `json:"params"`
	Permissions []string       `json:"permissions,omitempty" schema:"items.enum=read|write"`
	Code        string         `json:"code"`
	Detail      string         `json:"detail"`
}

func configure(e *contract.Extensions) error {
	if err := e.SecurityScheme("session", map[string]any{"type": "apiKey", "in": "cookie", "name": "session"}); err != nil {
		return err
	}
	if err := e.Security("createProduct", []map[string][]string{{"session": {}}}); err != nil {
		return err
	}
	if err := e.ResponseHeader("createProduct", "ETag", "Version", map[string]any{"type": "string"}); err != nil {
		return err
	}
	if err := e.ResponseHeader("createProduct", "Set-Cookie", "Session", map[string]any{"type": "string"}); err != nil {
		return err
	}
	if err := e.ErrorResponse("createProduct", 422, "Invalid fields", "application/problem+json", contract.TypeOf[Problem]()); err != nil {
		return err
	}
	return e.Extension("createProduct", "x-permission", "product.create")
}
func TestDocumentAndManifest(t *testing.T) {
	op := operation(t)
	for _, version := range []string{"3.0.0", "3.1.0"} {
		for _, format := range []string{"json", "yaml"} {
			t.Run(version+format, func(t *testing.T) {
				artifact, err := contract.Generate([]*contract.Operation{op}, contract.Options{Version: version, Format: format}, configure)
				require.NoError(t, err)
				again, err := contract.Generate([]*contract.Operation{op}, contract.Options{Version: version, Format: format}, configure)
				require.NoError(t, err)
				require.Equal(t, artifact, again)
				require.NoError(t, contract.CheckArtifact(artifact.Document, artifact.Manifest, []*contract.Operation{op}))
				require.Error(t, contract.CheckArtifact(append(artifact.Document, ' '), artifact.Manifest, []*contract.Operation{op}))
				op.Status = 202
				require.Error(t, contract.CheckArtifact(artifact.Document, artifact.Manifest, []*contract.Operation{op}))
				op.Status = 201
				require.Error(t, contract.CheckArtifact(artifact.Document, artifact.Manifest, nil))
				loader := openapi3.NewLoader()
				doc, err := loader.LoadFromData(artifact.Document)
				require.NoError(t, err)
				require.NoError(t, doc.Validate(context.Background()))
				require.Equal(t, []any{float64(1), float64(2)}, doc.Components.Schemas[op.Response].Value.Properties["level"].Value.Enum)
				route := doc.Paths.Find("/api/v1/products/{id}").Post
				require.Equal(t, "createProduct", route.OperationID)
				require.Contains(t, route.Responses.Value("201").Value.Headers, "Etag")
				require.NotNil(t, route.Responses.Value("422"))
				schema := route.RequestBody.Value.Content["application/json"].Schema.Value
				for _, body := range []string{`{"code":"ok","count":0,"enabled":false}`, `{"code":"ok","count":0,"enabled":false,"optional":null}`, `{"code":"ok","count":0,"enabled":false,"profile":{"name":"ok"}}`} {
					var v any
					require.NoError(t, json.Unmarshal([]byte(body), &v))
					require.NoError(t, schema.VisitJSON(v))
				}
				for _, body := range []string{`{"code":"ok","enabled":false}`, `{"code":"ok","count":0,"enabled":false,"profile":null}`, `{"code":"ok","count":0,"enabled":false,"remark":""}`} {
					var v any
					require.NoError(t, json.Unmarshal([]byte(body), &v))
					require.Error(t, schema.VisitJSON(v))
				}
			})
		}
	}
	require.Error(t, func() error {
		_, err := contract.Generate([]*contract.Operation{op, op}, contract.Options{}, nil)
		return err
	}())
}

type LegacyReq struct {
	mmeta.Meta `method:"GET" path:"/legacy"`
	Page       int `query:"page" binding:"omitempty,min=1"`
}
type LegacyRes struct{}
type DefaultReq struct {
	mmeta.Meta `method:"GET" path:"/default"`
	Page       int `query:"page" default:"0" binding:"min=1"`
}
type DefaultRes struct{}
type BoundReq struct {
	mmeta.Meta `method:"GET" path:"/bound"`
	Page       int `query:"page" binding:"min=1.5"`
}
type BoundRes struct{}

func TestUnsupportedAndInvalidDeclarations(t *testing.T) {
	_, err := contract.Compile(contract.TypeOf[LegacyReq](), contract.TypeOf[LegacyRes]())
	require.ErrorContains(t, err, "omitempty")
	_, err = contract.Compile(contract.TypeOf[DefaultReq](), contract.TypeOf[DefaultRes]())
	require.ErrorContains(t, err, "invalid default")
	_, err = contract.Compile(contract.TypeOf[BoundReq](), contract.TypeOf[BoundRes]())
	require.ErrorContains(t, err, "argument")
}

type CollisionReq struct {
	mmeta.Meta `method:"GET" path:"/collision"`
}
type CollisionRes struct {
	First  *first.Item  `json:"first"`
	Second *second.Item `json:"second"`
}
type Recursive struct {
	Name string     `json:"name"`
	Next *Recursive `json:"next,omitempty"`
}
type RecursiveReq struct {
	mmeta.Meta `method:"POST" path:"/recursive"`
	Value      *Recursive `json:"value"`
}
type RecursiveRes struct {
	Value *Recursive `json:"value,omitempty"`
}

func TestTypeIdentityAndRecursiveShapes(t *testing.T) {
	op, err := contract.Compile(contract.TypeOf[CollisionReq](), contract.TypeOf[CollisionRes]())
	require.NoError(t, err)
	artifact, err := contract.Generate([]*contract.Operation{op}, contract.Options{Format: "json"}, nil)
	require.NoError(t, err)
	require.Contains(t, string(artifact.Document), `"name"`)
	require.Contains(t, string(artifact.Document), `"count"`)
	fields := op.Types[op.Response].Fields
	require.NotEqual(t, fields[0].Type, fields[1].Type)
	op, err = contract.Compile(contract.TypeOf[RecursiveReq](), contract.TypeOf[RecursiveRes]())
	require.NoError(t, err)
	_, err = contract.Generate([]*contract.Operation{op}, contract.Options{}, nil)
	require.NoError(t, err)
	req := httptest.NewRequest("POST", "/recursive", strings.NewReader(`{"value":{"name":"one","next":{"name":"two"}}}`))
	req.Header.Set("Content-Type", "application/json")
	var target RecursiveReq
	_, err = op.Bind(req, nil, &target)
	require.NoError(t, err)
	require.Equal(t, "two", target.Value.Next.Name)
}
func TestExtensionInputsAreCopied(t *testing.T) {
	op := operation(t)
	artifact, err := contract.Generate([]*contract.Operation{op}, contract.Options{Format: "json"}, func(e *contract.Extensions) error {
		schema := map[string]any{"type": "string"}
		require.NoError(t, e.ResponseHeader("createProduct", "ETag", "Version", schema))
		schema["type"] = "integer"
		require.Error(t, e.Extension("createProduct", "responses", map[string]any{}))
		require.Error(t, e.ErrorResponse("createProduct", 201, "Success", "application/json", contract.TypeOf[Problem]()))
		require.Error(t, e.ResponseHeader("createProduct", "etag", "Duplicate", schema))
		require.Error(t, e.Extension("unknown", "x-a", true))
		return nil
	})
	require.NoError(t, err)
	doc, err := openapi3.NewLoader().LoadFromData(artifact.Document)
	require.NoError(t, err)
	require.True(t, doc.Paths.Find("/api/v1/products/{id}").Post.Responses.Value("201").Value.Headers["Etag"].Value.Schema.Value.Type.Is("string"))
}

type AllowReq struct {
	mmeta.Meta `method:"POST" path:"/allow" unknown:"allow"`
	Child      *Child `json:"child"`
}
type AllowRes struct{}
type StrictReq struct {
	mmeta.Meta `method:"POST" path:"/strict"`
	Child      *Child `json:"child"`
}
type StrictRes struct{}

func TestSharedTypeWithDifferentUnknownPolicies(t *testing.T) {
	allow, err := contract.Compile(contract.TypeOf[AllowReq](), contract.TypeOf[AllowRes]())
	require.NoError(t, err)
	strict, err := contract.Compile(contract.TypeOf[StrictReq](), contract.TypeOf[StrictRes]())
	require.NoError(t, err)
	a, err := contract.Generate([]*contract.Operation{allow, strict}, contract.Options{Format: "json"}, nil)
	require.NoError(t, err)
	doc, err := openapi3.NewLoader().LoadFromData(a.Document)
	require.NoError(t, err)
	value := map[string]any{"child": map[string]any{"name": "ok", "extra": true}}
	require.NoError(t, doc.Paths.Find("/allow").Post.RequestBody.Value.Content["application/json"].Schema.Value.VisitJSON(value))
	require.Error(t, doc.Paths.Find("/strict").Post.RequestBody.Value.Content["application/json"].Schema.Value.VisitJSON(value))
}

type Page[T any] struct {
	Items []T `json:"items"`
}
type GenericReq struct {
	mmeta.Meta `method:"GET" path:"/generic"`
}
type GenericRes struct {
	Page Page[Child] `json:"page"`
}

func TestGenericComponentName(t *testing.T) {
	op, err := contract.Compile(contract.TypeOf[GenericReq](), contract.TypeOf[GenericRes]())
	require.NoError(t, err)
	_, err = contract.Generate([]*contract.Operation{op}, contract.Options{}, nil)
	require.NoError(t, err)
}
