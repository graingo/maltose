package contract

import (
	"fmt"
	"mime"
	"net/http"
	"reflect"
	"strings"
)

// compileTransport keeps all stream declarations in the canonical operation.
func compileTransport(op *Operation, meta reflect.StructTag) error {
	requestStream := streamFieldIndex(op.Types[op.Request]) >= 0
	responseStream := streamFieldIndex(op.Types[op.Response]) >= 0
	if requestStream != (meta.Get("body") == "raw") {
		return fmt.Errorf("raw request requires body:raw and one stream:body io.ReadCloser field")
	}
	if responseStream != (meta.Get("response") == "raw") {
		return fmt.Errorf("raw response requires response:raw and one stream:body io.ReadCloser field")
	}
	if mode := meta.Get("response"); mode != "" && mode != "raw" {
		return fmt.Errorf("response must be raw or omitted for JSON")
	}
	if responseStream && (len(op.Types[op.Response].Fields) != 1 || op.Envelope != "none" || op.Status == 204 || op.Method == "HEAD") {
		return fmt.Errorf("raw response requires one stream field, envelope:none and a body-bearing success response")
	}
	var err error
	op.RawRequestMedia, err = streamMedia(meta.Get("consumes"), requestStream)
	if err != nil {
		return fmt.Errorf("consumes: %w", err)
	}
	op.RawResponseMedia, err = streamMedia(meta.Get("produces"), responseStream)
	if err != nil {
		return fmt.Errorf("produces: %w", err)
	}
	return nil
}
func streamMedia(raw string, enabled bool) (string, error) {
	if !enabled {
		if raw != "" {
			return "", fmt.Errorf("media declaration requires raw transport")
		}
		return "", nil
	}
	if raw == "" {
		return "application/octet-stream", nil
	}
	media, params, err := mime.ParseMediaType(raw)
	if err != nil || len(params) > 0 || !strings.Contains(media, "/") || strings.Contains(media, "*") {
		return "", fmt.Errorf("raw media type must be a concrete type without parameters: %q", raw)
	}
	return media, nil
}
func streamFieldIndex(t *Type) int {
	for i, f := range t.Fields {
		if f.Source == "stream" {
			return i
		}
	}
	return -1
}
func binarySchema() map[string]any { return map[string]any{"type": "string", "format": "binary"} }

// bindStream assigns the original resource. Presence follows HTTP framing and
// Content-Type, so a typed zero-length file is present without reading a byte.
func (op *Operation) bindStream(r *http.Request, target reflect.Value, input Input) error {
	f := op.Types[op.Request].Fields[streamFieldIndex(op.Types[op.Request])]
	present := r.Header.Get("Content-Type") != "" || r.ContentLength != 0 || len(r.TransferEncoding) > 0
	input["stream.body"] = Missing
	if present {
		input["stream.body"] = Present
		media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || media != op.RawRequestMedia {
			return &DecodeError{http.StatusUnsupportedMediaType, fmt.Errorf("Content-Type must be %s", op.RawRequestMedia)}
		}
	} else if f.Required {
		return invalid("stream.body", "required", "")
	}
	body := r.Body
	if body == nil {
		body = http.NoBody
	}
	target.FieldByIndex(f.Index).Set(reflect.ValueOf(body))
	return nil
}
