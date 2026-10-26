package contract_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/graingo/maltose/examples/raw-stream/api"
	"github.com/graingo/maltose/net/mhttp/contract"
	"github.com/graingo/maltose/util/mmeta"
	"github.com/stretchr/testify/require"
)

type DuplexReq struct {
	mmeta.Meta `method:"POST" path:"/streams/:id" body:"raw" response:"raw" produces:"image/png" status:"201" operation_id:"duplex"`
	ID         string        `path:"id" binding:"min=2"`
	Page       int           `query:"page" binding:"min=1"`
	Token      string        `header:"X-Token" field:"required"`
	Body       io.ReadCloser `stream:"body"`
}
type DuplexRes struct {
	Body io.ReadCloser `stream:"body"`
}

type trackedBody struct {
	io.Reader
	reads, closes int
}

func (b *trackedBody) Read(p []byte) (int, error) { b.reads++; return b.Reader.Read(p) }
func (b *trackedBody) Close() error               { b.closes++; return nil }

func TestRawBindingPreservesResource(t *testing.T) {
	op, err := contract.Compile(contract.TypeOf[DuplexReq](), contract.TypeOf[DuplexRes]())
	require.NoError(t, err)
	raw := []byte{0xff, 0, 0xfe, 1, 2}
	for _, tc := range []struct {
		name, media, id, page, token string
		want                         int
	}{
		{"bytes", "application/octet-stream", "ok", "1", "token", 0},
		{"media parameters", "application/octet-stream; charset=binary", "ok", "1", "token", 0},
		{"media mismatch", "application/json", "ok", "1", "token", 415},
		{"missing media", "", "ok", "1", "token", 415},
		{"path", "application/octet-stream", "x", "1", "token", 400},
		{"query", "application/octet-stream", "ok", "0", "token", 400},
		{"header", "application/octet-stream", "ok", "1", "", 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := &trackedBody{Reader: bytes.NewReader(raw)}
			r := httptest.NewRequest("POST", "/streams/ok?page="+tc.page, body)
			r.Header.Set("Content-Type", tc.media)
			if tc.token != "" {
				r.Header.Set("X-Token", tc.token)
			}
			var target DuplexReq
			input, err := op.Bind(r, map[string]string{"id": tc.id}, &target)
			require.Zero(t, body.reads)
			require.Zero(t, body.closes)
			if tc.want == 415 {
				var e *contract.DecodeError
				require.ErrorAs(t, err, &e)
				require.Equal(t, 415, e.Status)
				return
			}
			if tc.want == 400 {
				var e *contract.ValidationError
				require.ErrorAs(t, err, &e)
				return
			}
			require.NoError(t, err)
			require.Same(t, body, target.Body)
			require.Equal(t, contract.Present, input.Presence("stream.body"))
			got, err := io.ReadAll(target.Body)
			require.NoError(t, err)
			require.Equal(t, raw, got)
		})
	}
	// Optional absence differs from a declared empty file, without inspecting bytes.
	r := httptest.NewRequest("POST", "/streams/ok", nil)
	r.Header.Set("X-Token", "token")
	var target DuplexReq
	input, err := op.Bind(r, map[string]string{"id": "ok"}, &target)
	require.NoError(t, err)
	require.Equal(t, contract.Missing, input.Presence("stream.body"))
	require.Equal(t, http.NoBody, target.Body)
	// Size-limit middleware may wrap NoBody. Framing still represents absence.
	r.Body = http.MaxBytesReader(httptest.NewRecorder(), r.Body, 1024)
	input, err = op.Bind(r, map[string]string{"id": "ok"}, &target)
	require.NoError(t, err)
	require.Equal(t, contract.Missing, input.Presence("stream.body"))
	require.Same(t, r.Body, target.Body)
	upload, err := contract.Compile(contract.TypeOf[api.UploadReq](), contract.TypeOf[api.UploadRes]())
	require.NoError(t, err)
	for _, media := range []string{"", "application/octet-stream"} {
		r := httptest.NewRequest("PUT", "/files/id", nil)
		r.Header.Set("Content-Type", media)
		var req api.UploadReq
		input, err := upload.Bind(r, map[string]string{"id": "ccba9e91-bfa0-468e-91ac-956ee9097271"}, &req)
		if media == "" {
			require.ErrorContains(t, err, "stream.body: required")
		} else {
			require.NoError(t, err)
			require.Equal(t, contract.Present, input.Presence("stream.body"))
		}
	}
}

type streamErrorBody struct {
	Error string `json:"error"`
}

func TestRawOpenAPIAndManifest(t *testing.T) {
	upload, err := contract.Compile(contract.TypeOf[api.UploadReq](), contract.TypeOf[api.UploadRes]())
	require.NoError(t, err)
	download, err := contract.Compile(contract.TypeOf[api.DownloadReq](), contract.TypeOf[api.DownloadRes]())
	require.NoError(t, err)
	duplex, err := contract.Compile(contract.TypeOf[DuplexReq](), contract.TypeOf[DuplexRes]())
	require.NoError(t, err)
	operations := []*contract.Operation{upload, download, duplex}
	configure := func(e *contract.Extensions) error {
		if err := e.ResponseHeader("file.download", "Content-Disposition", "Download filename", map[string]any{"type": "string"}); err != nil {
			return err
		}
		if err := e.Extension("file.download", "x-audit", true); err != nil {
			return err
		}
		return e.ErrorResponse("file.download", 404, "Missing file", "application/json", contract.TypeOf[streamErrorBody]())
	}
	for _, version := range []string{"3.0.0", "3.1.0"} {
		for _, format := range []string{"json", "yaml"} {
			t.Run(version+format, func(t *testing.T) {
				opts := contract.Options{Version: version, Format: format}
				a, err := contract.Generate(operations, opts, configure)
				require.NoError(t, err)
				b, err := contract.Generate([]*contract.Operation{duplex, download, upload}, opts, configure)
				require.NoError(t, err)
				require.Equal(t, a, b)
				require.NoError(t, contract.CheckArtifact(a.Document, a.Manifest, operations))
				if format == "json" {
					var doc map[string]any
					require.NoError(t, json.Unmarshal(a.Document, &doc))
					paths := doc["paths"].(map[string]any)
					files := paths["/files/{id}"].(map[string]any)
					body := files["put"].(map[string]any)["requestBody"].(map[string]any)
					require.Equal(t, true, body["required"])
					checkBinarySchema(t, body, "application/octet-stream")
					get := files["get"].(map[string]any)
					responses := get["responses"].(map[string]any)
					success := responses["200"].(map[string]any)
					checkBinarySchema(t, success, "application/octet-stream")
					require.Contains(t, success["headers"], "Content-Disposition")
					require.Contains(t, responses, "404")
					require.Equal(t, true, get["x-audit"])
					both := paths["/streams/{id}"].(map[string]any)["post"].(map[string]any)
					require.Len(t, both["parameters"], 3)
					require.Equal(t, false, both["requestBody"].(map[string]any)["required"])
					checkBinarySchema(t, both["responses"].(map[string]any)["201"].(map[string]any), "image/png")
				}
				for _, mutate := range []func(*contract.Operation){func(o *contract.Operation) { o.Body = "none" }, func(o *contract.Operation) { o.RawRequestMedia = "image/png" }} {
					changed := *upload
					mutate(&changed)
					require.ErrorContains(t, contract.CheckArtifact(a.Document, a.Manifest, []*contract.Operation{&changed, download, duplex}), "contract differs")
				}
				changed := *download
				changed.RawResponseMedia = "image/png"
				require.ErrorContains(t, contract.CheckArtifact(a.Document, a.Manifest, []*contract.Operation{upload, &changed, duplex}), "contract differs")
			})
		}
	}
}
func checkBinarySchema(t *testing.T, body map[string]any, media string) {
	t.Helper()
	content := body["content"].(map[string]any)
	require.Len(t, content, 1)
	require.Equal(t, map[string]any{"type": "string", "format": "binary"}, content[media].(map[string]any)["schema"])
}

type MissingFieldReq struct {
	mmeta.Meta `method:"POST" path:"/bad" body:"raw"`
}
type MissingFieldRes struct{}

type MissingModeReq struct {
	mmeta.Meta `method:"POST" path:"/bad" `
	Body       io.ReadCloser `stream:"body"`
}
type MissingModeRes struct{}

type JSONConflictReq struct {
	mmeta.Meta `method:"POST" path:"/bad" body:"raw"`
	Body       io.ReadCloser `stream:"body"`
	Value      string        `json:"value"`
}
type JSONConflictRes struct{}

type FormConflictReq struct {
	mmeta.Meta `method:"POST" path:"/bad" body:"raw"`
	Body       io.ReadCloser `stream:"body"`
	Value      string        `form:"value"`
}
type FormConflictRes struct{}

type ReaderTypeReq struct {
	mmeta.Meta `method:"POST" path:"/bad" body:"raw"`
	Body       io.Reader `stream:"body"`
}
type ReaderTypeRes struct{}

type NullableReq struct {
	mmeta.Meta `method:"POST" path:"/bad" body:"raw"`
	Body       io.ReadCloser `stream:"body" field:"nullable"`
}
type NullableRes struct{}

type BindingReq struct {
	mmeta.Meta `method:"POST" path:"/bad" body:"raw"`
	Body       io.ReadCloser `stream:"body" binding:"required"`
}
type BindingRes struct{}

type JSONTagReq struct {
	mmeta.Meta `method:"POST" path:"/bad" body:"raw"`
	Body       io.ReadCloser `stream:"body" json:"-"`
}
type JSONTagRes struct{}

type DuplicateReq struct {
	mmeta.Meta `method:"POST" path:"/bad" body:"raw"`
	Body       io.ReadCloser `stream:"body"`
	Other      io.ReadCloser `stream:"body"`
}
type DuplicateRes struct{}

type NestedReq struct {
	mmeta.Meta `method:"POST" path:"/bad" body:"raw"`
	Value      struct {
		Body io.ReadCloser `stream:"body"`
	} `json:"value"`
}
type NestedRes struct{}

type OutputMissingReq struct {
	mmeta.Meta `method:"POST" path:"/bad" response:"raw"`
}
type OutputMissingRes struct{}

type OutputModeReq struct {
	mmeta.Meta `method:"POST" path:"/bad" `
}
type OutputModeRes struct {
	Body io.ReadCloser `stream:"body"`
}

type EnvelopeReq struct {
	mmeta.Meta `method:"POST" path:"/bad" response:"raw" envelope:"maltose"`
}
type EnvelopeRes struct {
	Body io.ReadCloser `stream:"body"`
}

type NoContentReq struct {
	mmeta.Meta `method:"POST" path:"/bad" response:"raw" status:"204"`
}
type NoContentRes struct {
	Body io.ReadCloser `stream:"body"`
}

type OutputJSONReq struct {
	mmeta.Meta `method:"POST" path:"/bad" response:"raw"`
}
type OutputJSONRes struct {
	Body  io.ReadCloser `stream:"body"`
	Value string        `json:"value"`
}

type MediaWithoutModeReq struct {
	mmeta.Meta `method:"POST" path:"/bad" consumes:"image/png"`
}
type MediaWithoutModeRes struct{}

type MediaWildcardReq struct {
	mmeta.Meta `method:"POST" path:"/bad" body:"raw" consumes:"image/*"`
	Body       io.ReadCloser `stream:"body"`
}
type MediaWildcardRes struct{}

type MediaParametersReq struct {
	mmeta.Meta `method:"POST" path:"/bad" response:"raw" produces:"text/plain; charset=utf-8"`
}
type MediaParametersRes struct {
	Body io.ReadCloser `stream:"body"`
}

type ResponseModeReq struct {
	mmeta.Meta `method:"POST" path:"/bad" response:"stream"`
}
type ResponseModeRes struct{}

func TestInvalidRawContracts(t *testing.T) {
	t.Run("MissingField", func(t *testing.T) {
		_, err := contract.Compile(contract.TypeOf[MissingFieldReq](), contract.TypeOf[MissingFieldRes]())
		require.ErrorContains(t, err, "raw request requires")
	})
	t.Run("MissingMode", func(t *testing.T) {
		_, err := contract.Compile(contract.TypeOf[MissingModeReq](), contract.TypeOf[MissingModeRes]())
		require.ErrorContains(t, err, "raw request requires")
	})
	t.Run("JSONConflict", func(t *testing.T) {
		_, err := contract.Compile(contract.TypeOf[JSONConflictReq](), contract.TypeOf[JSONConflictRes]())
		require.ErrorContains(t, err, "path/query/header")
	})
	t.Run("FormConflict", func(t *testing.T) {
		_, err := contract.Compile(contract.TypeOf[FormConflictReq](), contract.TypeOf[FormConflictRes]())
		require.ErrorContains(t, err, "path/query/header")
	})
	t.Run("ReaderType", func(t *testing.T) {
		_, err := contract.Compile(contract.TypeOf[ReaderTypeReq](), contract.TypeOf[ReaderTypeRes]())
		require.ErrorContains(t, err, "requires io.ReadCloser")
	})
	t.Run("Nullable", func(t *testing.T) {
		_, err := contract.Compile(contract.TypeOf[NullableReq](), contract.TypeOf[NullableRes]())
		require.ErrorContains(t, err, "requires io.ReadCloser")
	})
	t.Run("Binding", func(t *testing.T) {
		_, err := contract.Compile(contract.TypeOf[BindingReq](), contract.TypeOf[BindingRes]())
		require.ErrorContains(t, err, "conflicts with binding")
	})
	t.Run("JSONTag", func(t *testing.T) {
		_, err := contract.Compile(contract.TypeOf[JSONTagReq](), contract.TypeOf[JSONTagRes]())
		require.ErrorContains(t, err, "conflicts with json")
	})
	t.Run("Duplicate", func(t *testing.T) {
		_, err := contract.Compile(contract.TypeOf[DuplicateReq](), contract.TypeOf[DuplicateRes]())
		require.ErrorContains(t, err, "duplicate field")
	})
	t.Run("Nested", func(t *testing.T) {
		_, err := contract.Compile(contract.TypeOf[NestedReq](), contract.TypeOf[NestedRes]())
		require.ErrorContains(t, err, "direct root")
	})
	t.Run("OutputMissing", func(t *testing.T) {
		_, err := contract.Compile(contract.TypeOf[OutputMissingReq](), contract.TypeOf[OutputMissingRes]())
		require.ErrorContains(t, err, "raw response requires")
	})
	t.Run("OutputMode", func(t *testing.T) {
		_, err := contract.Compile(contract.TypeOf[OutputModeReq](), contract.TypeOf[OutputModeRes]())
		require.ErrorContains(t, err, "raw response requires")
	})
	t.Run("Envelope", func(t *testing.T) {
		_, err := contract.Compile(contract.TypeOf[EnvelopeReq](), contract.TypeOf[EnvelopeRes]())
		require.ErrorContains(t, err, "envelope:none")
	})
	t.Run("NoContent", func(t *testing.T) {
		_, err := contract.Compile(contract.TypeOf[NoContentReq](), contract.TypeOf[NoContentRes]())
		require.ErrorContains(t, err, "body-bearing")
	})
	t.Run("OutputJSON", func(t *testing.T) {
		_, err := contract.Compile(contract.TypeOf[OutputJSONReq](), contract.TypeOf[OutputJSONRes]())
		require.ErrorContains(t, err, "one stream field")
	})
	t.Run("MediaWithoutMode", func(t *testing.T) {
		_, err := contract.Compile(contract.TypeOf[MediaWithoutModeReq](), contract.TypeOf[MediaWithoutModeRes]())
		require.ErrorContains(t, err, "requires raw transport")
	})
	t.Run("MediaWildcard", func(t *testing.T) {
		_, err := contract.Compile(contract.TypeOf[MediaWildcardReq](), contract.TypeOf[MediaWildcardRes]())
		require.ErrorContains(t, err, "concrete type")
	})
	t.Run("MediaParameters", func(t *testing.T) {
		_, err := contract.Compile(contract.TypeOf[MediaParametersReq](), contract.TypeOf[MediaParametersRes]())
		require.ErrorContains(t, err, "concrete type")
	})
	t.Run("ResponseMode", func(t *testing.T) {
		_, err := contract.Compile(contract.TypeOf[ResponseModeReq](), contract.TypeOf[ResponseModeRes]())
		require.ErrorContains(t, err, "response must")
	})
}

type RawHeadReq struct {
	mmeta.Meta `method:"HEAD" path:"/raw" response:"raw"`
}
type RawHeadRes struct {
	Body io.ReadCloser `stream:"body"`
}
type RawNoContentReq struct {
	mmeta.Meta `method:"POST" path:"/raw" body:"raw" status:"204"`
	Body       io.ReadCloser `stream:"body"`
}
type RawNoContentRes struct{}

func TestRawBodyBearingResponses(t *testing.T) {
	_, err := contract.Compile(contract.TypeOf[RawHeadReq](), contract.TypeOf[RawHeadRes]())
	require.ErrorContains(t, err, "body-bearing")
	op, err := contract.Compile(contract.TypeOf[RawNoContentReq](), contract.TypeOf[RawNoContentRes]())
	require.NoError(t, err)
	artifact, err := contract.Generate([]*contract.Operation{op}, contract.Options{Format: "json"}, nil)
	require.NoError(t, err)
	var doc map[string]any
	require.NoError(t, json.Unmarshal(artifact.Document, &doc))
	operation := doc["paths"].(map[string]any)["/raw"].(map[string]any)["post"].(map[string]any)
	require.Contains(t, operation, "requestBody")
	require.NotContains(t, operation["responses"].(map[string]any)["204"], "content")
}
