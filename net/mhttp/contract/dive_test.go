package contract_test

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/graingo/maltose/net/mhttp/contract"
	"github.com/graingo/maltose/util/mmeta"
	"github.com/stretchr/testify/require"
)

type DiveReq struct {
	mmeta.Meta `method:"POST" path:"/dive"`
	Slices     []*[]string             `json:"slices" binding:"dive,dive,required"`
	Maps       map[string]*[]string    `json:"maps" binding:"dive,dive,required"`
	Arrays     [1]*[1]string           `json:"arrays" binding:"dive,dive,required"`
	Required   []*[]string             `json:"required" binding:"dive,required,dive,required"`
	Enum       []*string               `json:"enum" binding:"dive,oneof=a b"`
	Deep       []*map[string]*[]string `json:"deep" binding:"dive,dive,dive,required"`
}

type DiveRes struct{}

func TestDiveRuleScope(t *testing.T) {
	op, err := contract.Compile(contract.TypeOf[DiveReq](), contract.TypeOf[DiveRes]())
	require.NoError(t, err)
	for _, version := range []string{"3.0.0", "3.1.0"} {
		t.Run(version, func(t *testing.T) {
			artifact, err := contract.Generate([]*contract.Operation{op}, contract.Options{Version: version, Format: "json"}, nil)
			require.NoError(t, err)
			doc, err := openapi3.NewLoader().LoadFromData(artifact.Document)
			require.NoError(t, err)
			schema := doc.Paths.Find("/dive").Post.RequestBody.Value.Content["application/json"].Schema.Value
			// The enum belongs to the non-null branch. Some OpenAPI validators
			// short-circuit on null, so also verify the exported union's scope.
			enumItem := schema.Properties["enum"].Value.Items.Value
			require.Empty(t, enumItem.Enum)
			require.Len(t, enumItem.AnyOf, 2)
			require.Equal(t, []any{"a", "b"}, enumItem.AnyOf[0].Value.Enum)
			for _, tc := range []struct {
				body, field, code string
			}{
				{`{}`, "", ""},
				{`{"slices":[null,[],["ok"]]}`, "", ""},
				{`{"slices":[[""]]}`, "json.slices[0][0]", "required"},
				{`{"maps":{"a":null,"b":[],"c":["ok"]}}`, "", ""},
				{`{"maps":{"a":[""]}}`, `json.maps["a"][0]`, "required"},
				{`{"arrays":[null]}`, "", ""},
				{`{"arrays":[["ok"]]}`, "", ""},
				{`{"arrays":[[""]]}`, "json.arrays[0][0]", "required"},
				{`{"required":[["ok"]]}`, "", ""},
				{`{"required":[null]}`, "json.required[0]", "required"},
				{`{"required":[[]]}`, "json.required[0]", "required"},
				{`{"required":[[""]]}`, "json.required[0][0]", "required"},
				{`{"enum":[null,"a","b"]}`, "", ""},
				{`{"enum":["c"]}`, "json.enum[0]", "oneof"},
				{`{"deep":[null,{}, {"a":null,"b":[],"c":["ok"]}]}`, "", ""},
				{`{"deep":[{"a":[""]}]}`, `json.deep[0]["a"][0]`, "required"},
			} {
				t.Run(tc.body, func(t *testing.T) {
					request := httptest.NewRequest("POST", "/dive", strings.NewReader(tc.body))
					request.Header.Set("Content-Type", "application/json")
					var target DiveReq
					_, bindErr := op.Bind(request, nil, &target)
					var value any
					require.NoError(t, json.Unmarshal([]byte(tc.body), &value))
					schemaErr := schema.VisitJSON(value)
					if tc.code == "" {
						require.NoError(t, bindErr)
						require.NoError(t, schemaErr)
						return
					}
					var validation *contract.ValidationError
					require.ErrorAs(t, bindErr, &validation)
					require.Equal(t, tc.field, validation.Issues[0].Field)
					require.Equal(t, tc.code, validation.Issues[0].Code)
					require.Error(t, schemaErr)
				})
			}
		})
	}
}
