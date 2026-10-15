package contract_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/graingo/maltose/net/mhttp/contract"
	"github.com/graingo/maltose/util/mmeta"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/require"
)

type ExampleReq struct {
	mmeta.Meta `method:"GET" path:"/example"`
}
type ExampleRes struct {
	PermissionCodes []string `json:"permission_codes" field:"required"`
}
type ResponseChild struct {
	Items []string `json:"items" field:"required"`
}
type ResponseInline struct {
	Flag bool `json:"flag" field:"required"`
}
type ResponseReq struct {
	mmeta.Meta `method:"GET" path:"/response" envelope:"maltose"`
}
type ResponseRes struct {
	ResponseInline    `field:"inline"`
	Slice             []string          `json:"slice" field:"required"`
	Map               map[string]string `json:"map" field:"required"`
	Pointer           *string           `json:"pointer" field:"required"`
	PointerSlice      **[]string        `json:"pointer_slice" field:"required"`
	Any               any               `json:"any" field:"required"`
	Nullable          []string          `json:"nullable" field:"required,nullable"`
	NullableMap       map[string]string `json:"nullable_map" field:"required,nullable"`
	NullablePointer   *ResponseChild    `json:"nullable_pointer" field:"required,nullable"`
	PlainSlice        []string          `json:"plain_slice"`
	PlainMap          map[string]string `json:"plain_map"`
	PlainPointer      *string           `json:"plain_pointer"`
	PlainAny          any               `json:"plain_any"`
	OmitSlice         []string          `json:"omit_slice,omitempty"`
	OmitMap           map[string]string `json:"omit_map,omitempty"`
	OmitPointer       *string           `json:"omit_pointer,omitempty"`
	OmitPointerSlice  *[]string         `json:"omit_pointer_slice,omitempty"`
	OmitDoublePointer **string          `json:"omit_double_pointer,omitempty"`
	OmitAny           any               `json:"omit_any,omitempty"`
	OmitStruct        ResponseChild     `json:"omit_struct,omitempty" field:"required"`
	OmitArray         [1]string         `json:"omit_array,omitempty" field:"required"`
	OmitEmptyArray    [0]string         `json:"omit_empty_array,omitempty"`
	ExplicitNullable  []string          `json:"explicit_nullable,omitempty" field:"nullable"`
	Elements          []*string         `json:"elements" field:"required"`
	Nested            []*ResponseChild  `json:"nested" field:"required"`
}

func TestResponseFieldContract(t *testing.T) {
	op, err := contract.Compile(contract.TypeOf[ResponseReq](), contract.TypeOf[ResponseRes]())
	require.NoError(t, err)
	cases := map[string]struct {
		required, nullable bool
		value              any
	}{
		"flag":                {true, false, false},
		"slice":               {true, false, []any{}},
		"map":                 {true, false, map[string]any{}},
		"pointer":             {true, false, ""},
		"pointer_slice":       {true, false, []any{}},
		"any":                 {true, false, false},
		"nullable":            {true, true, []any{}},
		"nullable_map":        {true, true, map[string]any{}},
		"nullable_pointer":    {true, true, map[string]any{"items": []any{}}},
		"plain_slice":         {true, true, []any{}},
		"plain_map":           {true, true, map[string]any{}},
		"plain_pointer":       {true, true, ""},
		"plain_any":           {true, true, false},
		"omit_slice":          {false, false, []any{}},
		"omit_map":            {false, false, map[string]any{}},
		"omit_pointer":        {false, false, ""},
		"omit_pointer_slice":  {false, true, []any{}},
		"omit_double_pointer": {false, true, ""},
		"omit_any":            {false, true, false},
		"omit_struct":         {true, false, map[string]any{"items": []any{}}},
		"omit_array":          {true, false, []any{""}},
		"omit_empty_array":    {false, false, []any{}},
		"explicit_nullable":   {false, true, []any{}},
		"elements":            {true, false, []any{nil, ""}},
		"nested":              {true, false, []any{nil, map[string]any{"items": []any{}}}},
	}
	require.Len(t, op.Types[op.Response].Fields, len(cases))
	for _, f := range op.Types[op.Response].Fields {
		tc := cases[f.Name]
		require.Equal(t, tc.required, f.Required, f.Name)
		require.Equal(t, tc.nullable, f.Nullable, f.Name)
	}
	for _, version := range []string{"3.0.0", "3.1.0"} {
		t.Run(version, func(t *testing.T) {
			artifact, err := contract.Generate([]*contract.Operation{op}, contract.Options{Version: version, Format: "json"}, nil)
			require.NoError(t, err)
			doc, err := openapi3.NewLoader().LoadFromData(artifact.Document)
			require.NoError(t, err)
			require.NoError(t, doc.Validate(context.Background()))
			schema := doc.Components.Schemas[op.Response].Value
			validate := func(value any) error { return schema.VisitJSON(value) }
			if version == "3.1.0" {
				var document any
				require.NoError(t, json.Unmarshal(artifact.Document, &document))
				compiler := jsonschema.NewCompiler()
				compiler.DefaultDraft(jsonschema.Draft2020)
				require.NoError(t, compiler.AddResource("https://example.test/openapi.json", document))
				compiled, err := compiler.Compile("https://example.test/openapi.json#/components/schemas/" + op.Response)
				require.NoError(t, err)
				validate = compiled.Validate
			}
			body := map[string]any{}
			for name, tc := range cases {
				body[name] = tc.value
			}
			require.NoError(t, validate(body))
			for name, tc := range cases {
				t.Run(name, func(t *testing.T) {
					property := schema.Properties[name].Value
					require.NoError(t, property.VisitJSON(tc.value))
					if tc.nullable {
						require.NoError(t, property.VisitJSON(nil))
					} else {
						require.Error(t, property.VisitJSON(nil))
					}
					body[name] = nil
					if tc.nullable {
						require.NoError(t, validate(body))
					} else {
						require.Error(t, validate(body))
					}
					delete(body, name)
					if tc.required {
						require.Error(t, validate(body))
					} else {
						require.NoError(t, validate(body))
					}
					body[name] = tc.value
				})
			}
			require.Error(t, schema.Properties["nested"].Value.VisitJSON([]any{map[string]any{}}))
			require.Error(t, schema.Properties["nested"].Value.VisitJSON([]any{map[string]any{"items": nil}}))
			// The same compiled response is used inside the standard envelope.
			success := doc.Paths.Find("/response").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value
			require.NoError(t, success.VisitJSON(map[string]any{"code": float64(0), "message": "OK", "data": body}))
		})
	}
}

func TestResponsePresenceExampleAndManifest(t *testing.T) {
	compile := func() *contract.Operation {
		op, err := contract.Compile(contract.TypeOf[ExampleReq](), contract.TypeOf[ExampleRes]())
		require.NoError(t, err)
		return op
	}
	for _, version := range []string{"3.0.0", "3.1.0"} {
		for _, format := range []string{"json", "yaml"} {
			t.Run(version+format, func(t *testing.T) {
				op := compile()
				options := contract.Options{Version: version, Format: format}
				first, err := contract.Generate([]*contract.Operation{op}, options, nil)
				require.NoError(t, err)
				again, err := contract.Generate([]*contract.Operation{compile()}, options, nil)
				require.NoError(t, err)
				require.Equal(t, first.Document, again.Document)
				require.Equal(t, first.Manifest, again.Manifest)
				require.NoError(t, contract.CheckArtifact(first.Document, first.Manifest, []*contract.Operation{compile()}))
				doc, err := openapi3.NewLoader().LoadFromData(first.Document)
				require.NoError(t, err)
				schema := doc.Components.Schemas[op.Response].Value
				property := schema.Properties["permission_codes"].Value
				require.Equal(t, []string{"permission_codes"}, schema.Required)
				require.True(t, property.Type.Is("array"))
				require.Empty(t, property.AnyOf)
				require.True(t, property.Items.Value.Type.Is("string"))
				encoded, err := json.Marshal(ExampleRes{PermissionCodes: []string{}})
				require.NoError(t, err)
				require.JSONEq(t, `{"permission_codes":[]}`, string(encoded))
				for _, tc := range []struct {
					body  string
					valid bool
				}{
					{string(encoded), true}, {`{"permission_codes":["read"]}`, true},
					{`{}`, false}, {`{"permission_codes":null}`, false}, {`{"permission_codes":[null]}`, false},
				} {
					var value any
					require.NoError(t, json.Unmarshal([]byte(tc.body), &value))
					if tc.valid {
						require.NoError(t, schema.VisitJSON(value))
					} else {
						require.Error(t, schema.VisitJSON(value))
					}
				}
				// A changed null/presence promise invalidates the old manifest.
				for _, change := range []func(*contract.Field){func(f *contract.Field) { f.Nullable = true }, func(f *contract.Field) { f.Required = false }} {
					changed := compile()
					change(&changed.Types[changed.Response].Fields[0])
					require.NotEqual(t, op.Fingerprint(), changed.Fingerprint())
					require.Error(t, contract.CheckArtifact(first.Document, first.Manifest, []*contract.Operation{changed}))
				}
			})
		}
	}
}

type ResponseOmitReq struct {
	mmeta.Meta `method:"GET" path:"/omit"`
}
type ResponseOmitRes struct {
	Values []string `json:"values,omitempty" field:"required"`
}
type ResponseNullOmitReq struct {
	mmeta.Meta `method:"GET" path:"/omit-null"`
}
type ResponseNullOmitRes struct {
	Value *string `json:"value,omitempty" field:"required,nullable"`
}
type ResponseScalarReq struct {
	mmeta.Meta `method:"GET" path:"/scalar"`
}
type ResponseScalarRes struct {
	Value string `json:"value" field:"nullable"`
}

func TestResponseInvalidDeclarations(t *testing.T) {
	_, err := contract.Compile(contract.TypeOf[ResponseOmitReq](), contract.TypeOf[ResponseOmitRes]())
	require.ErrorContains(t, err, "required conflicts with json omitempty")
	_, err = contract.Compile(contract.TypeOf[ResponseNullOmitReq](), contract.TypeOf[ResponseNullOmitRes]())
	require.ErrorContains(t, err, "required conflicts with json omitempty")
	_, err = contract.Compile(contract.TypeOf[ResponseScalarReq](), contract.TypeOf[ResponseScalarRes]())
	require.ErrorContains(t, err, "nullable requires")
}
