package contract

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

type Options struct {
	Version string
	Title   string
	Format  string
}
type Artifact struct {
	Document []byte
	Manifest []byte
}
type Manifest struct {
	Version        int               `json:"version"`
	DocumentSHA256 string            `json:"document_sha256"`
	Operations     map[string]string `json:"operations"`
}

func digest(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }

// Generate exports the same graph executed by Bind. Configure can only add
// business-owned declarations through Extensions.
func Generate(operations []*Operation, options Options, configure func(*Extensions) error) (Artifact, error) {
	if err := ValidateOperations(operations); err != nil {
		return Artifact{}, err
	}
	if options.Version == "" {
		options.Version = "3.1.0"
	}
	if options.Version != "3.0.0" && options.Version != "3.1.0" {
		return Artifact{}, fmt.Errorf("OpenAPI version must be 3.0.0 or 3.1.0")
	}
	if options.Format == "" {
		options.Format = "yaml"
	}
	if options.Format != "yaml" && options.Format != "json" {
		return Artifact{}, fmt.Errorf("format must be yaml or json")
	}
	if options.Title == "" {
		options.Title = "API"
	}
	doc := map[string]any{"openapi": options.Version, "info": map[string]any{"title": options.Title, "version": "1.0.0"}, "paths": map[string]any{}, "components": map[string]any{"schemas": map[string]any{}, "securitySchemes": map[string]any{}}}
	components := doc["components"].(map[string]any)["schemas"].(map[string]any)
	ext := &Extensions{doc: doc, operations: map[string]map[string]any{}, statuses: map[string]int{}, version: options.Version, schemas: components}
	manifest := Manifest{Version: Version, Operations: map[string]string{}}
	operations = append([]*Operation(nil), operations...)
	sort.Slice(operations, func(i, j int) bool {
		return operations[i].Method+operations[i].Path < operations[j].Method+operations[j].Path
	})
	for _, op := range operations {
		key := op.Method + " " + op.Path

		manifest.Operations[key] = op.Fingerprint()
		b := schemaWriter{op: op, version: options.Version, components: components}
		operation, err := b.operation()
		if err != nil {
			return Artifact{}, fmt.Errorf("%s: %w", key, err)
		}
		paths := doc["paths"].(map[string]any)
		p := op.OpenAPIPath()
		if paths[p] == nil {
			paths[p] = map[string]any{}
		}
		paths[p].(map[string]any)[strings.ToLower(op.Method)] = operation
		ext.operations[op.ID] = operation
		ext.statuses[op.ID] = op.Status
	}
	if configure != nil {
		if err := configure(ext); err != nil {
			return Artifact{}, err
		}
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return Artifact{}, err
	}
	loader := openapi3.NewLoader()
	parsed, err := loader.LoadFromData(data)
	if err != nil {
		return Artifact{}, fmt.Errorf("parse generated OpenAPI: %w", err)
	}
	if err = parsed.Validate(context.Background()); err != nil {
		return Artifact{}, fmt.Errorf("validate generated OpenAPI: %w", err)
	}
	if options.Format == "yaml" {
		var node yaml.Node
		if err = yaml.Unmarshal(data, &node); err != nil {
			return Artifact{}, err
		}
		blockYAML(&node)
		data, err = yaml.Marshal(&node)
		if err != nil {
			return Artifact{}, err
		}
	} else {
		data = append(data, '\n')
	}
	manifest.DocumentSHA256 = digest(data)
	index, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return Artifact{}, err
	}
	return Artifact{data, append(index, '\n')}, nil
}

// CheckArtifact verifies the final bytes and exact registered route set. It does
// not attempt to prove business validators or middleware behavior from Schema.
func CheckArtifact(document, index []byte, operations []*Operation) error {
	var m Manifest
	if err := json.Unmarshal(index, &m); err != nil {
		return fmt.Errorf("parse contract manifest: %w", err)
	}
	if m.Version != Version {
		return fmt.Errorf("unsupported contract manifest version %d", m.Version)
	}
	if digest(document) != m.DocumentSHA256 {
		return fmt.Errorf("OpenAPI document digest differs from manifest")
	}
	if len(operations) != len(m.Operations) {
		return fmt.Errorf("registered operation count differs from manifest")
	}
	seen := map[string]bool{}
	for _, op := range operations {
		key := op.Method + " " + op.Path
		if seen[key] || m.Operations[key] != op.Fingerprint() {
			return fmt.Errorf("contract differs for %s; regenerate OpenAPI", key)
		}
		seen[key] = true
	}
	return nil
}

type schemaWriter struct {
	op         *Operation
	version    string
	components map[string]any
}

func (b schemaWriter) ref(id string, request bool) (map[string]any, error) {
	n := b.op.Types[id]
	if n.Kind == "object" {
		if _, exists := b.components[id]; !exists {
			b.components[id] = map[string]any{}
			schema, err := b.object(n, request, false)
			if err != nil {
				return nil, err
			}
			b.components[id] = schema
		}
		return map[string]any{"$ref": "#/components/schemas/" + id}, nil
	}
	if n.Kind == "pointer" {
		s, err := b.ref(n.Elem, request)
		if err != nil {
			return nil, err
		}
		return nullable(s, b.version), nil
	}
	s := map[string]any{}
	switch n.Kind {
	case "any":
		return s, nil
	case "integer":
		s["type"] = "integer"
		bits := n.Go.Bits()
		if n.Go.Kind() >= reflect.Uint && n.Go.Kind() <= reflect.Uint64 {
			s["minimum"] = uint64(0)
			s["maximum"] = ^uint64(0) >> (64 - bits)
		} else {
			s["minimum"] = -int64(1) << (bits - 1)
			s["maximum"] = int64(^uint64(0) >> (65 - bits))
		}
	case "time":
		s["type"] = "string"
		s["format"] = "date-time"
	case "bytes":
		s["type"] = "string"
		s["format"] = "byte"
	case "map":
		s["type"] = "object"
		child, err := b.ref(n.Elem, request)
		if err != nil {
			return nil, err
		}
		s["additionalProperties"] = child
	case "array":
		s["type"] = "array"
		child, err := b.ref(n.Elem, request)
		if err != nil {
			return nil, err
		}
		s["items"] = child
		if n.Go.Kind() == reflect.Array {
			s["minItems"] = n.Length
			s["maxItems"] = n.Length
		}
	default:
		s["type"] = n.Kind
	}
	if !request && (n.Go.Kind() == reflect.Slice || n.Go.Kind() == reflect.Map) {
		return nullable(s, b.version), nil
	}
	return s, nil
}
func nullable(s map[string]any, version string) map[string]any {
	if version == "3.1.0" {
		return map[string]any{"anyOf": []any{s, map[string]any{"type": "null"}}}
	}
	// OpenAPI 3.0 nullable only applies beside an explicit type. A null-only
	// branch permits null for both references and schemas with constraints.
	return map[string]any{"anyOf": []any{s, map[string]any{"type": "object", "nullable": true, "enum": []any{nil}}}}
}
func (b schemaWriter) field(f Field, request bool) (map[string]any, error) {
	id := f.Type
	if request {
		for b.op.Types[id].Kind == "pointer" {
			id = b.op.Types[id].Elem
		}
	}
	s, err := b.ref(id, request)
	if err != nil {
		return nil, err
	}
	if !request && f.OmitEmpty {
		if branches, ok := s["anyOf"].([]any); ok {
			s = branches[0].(map[string]any)
		}
	}
	if _, ok := s["$ref"]; ok {
		s = map[string]any{"allOf": []any{s}}
	}
	if f.Description != "" {
		s["description"] = f.Description
	}
	inferred := map[string]any{}
	if request {
		if err := b.rules(s, b.op.Types[f.Type].Go, f.Rules); err != nil {
			return nil, err
		}
		if err := b.rules(inferred, b.op.Types[f.Type].Go, f.Rules); err != nil {
			return nil, err
		}
	}
	if err := supplement(constraintTarget(s), f.Schema, inferred); err != nil {
		return nil, fmt.Errorf("%s: %w", f.Name, err)
	}
	if len(f.Default) > 0 {
		target := reflect.New(b.op.Types[f.Type].Go)
		if err := json.Unmarshal(f.Default, target.Interface()); err != nil {
			return nil, err
		}
		v := target.Elem().Interface()
		if existing, ok := s["default"]; ok {
			a, _ := json.Marshal(existing)
			c, _ := json.Marshal(v)
			if string(a) != string(c) {
				return nil, fmt.Errorf("schema default conflicts for %s", f.Name)
			}
		}
		s["default"] = v
	}
	if request && f.Nullable {
		s = nullable(s, b.version)
	}
	return s, nil
}
func (b schemaWriter) object(n *Type, request, root bool) (map[string]any, error) {
	properties := map[string]any{}
	required := []string{}
	for _, f := range n.Fields {
		if root && f.Source != "json" && f.Source != "form" {
			continue
		}
		s, err := b.field(f, request)
		if err != nil {
			return nil, err
		}
		properties[f.Name] = s
		if request && f.Required || !request && (!f.OmitEmpty || !canOmit(b.op.Types[f.Type].Go)) {
			required = append(required, f.Name)
		}
	}
	schema := map[string]any{"type": "object", "properties": properties}
	if request {
		schema["additionalProperties"] = b.op.Unknown == "allow"
	}
	if len(required) > 0 {
		sort.Strings(required)
		schema["required"] = required
	}
	return schema, nil
}
func (b schemaWriter) rules(s map[string]any, t reflect.Type, rules []Rule) error {
	// Pointer schemas separate a value from null. Apply value rules inside that
	// branch, and retain null only when this scope permits it.
	if branches, ok := s["anyOf"].([]any); ok {
		value := branches[0].(map[string]any)
		if err := b.rules(value, t, rules); err != nil {
			return err
		}
		if requiresValue(rules) {
			delete(s, "anyOf")
			s["allOf"] = []any{value}
		}
		return nil
	}
	t = indirect(t)
	for i, r := range rules {
		if r.Name == "dive" {
			key := "items"
			if t.Kind() == reflect.Map {
				key = "additionalProperties"
			}
			child, ok := s[key].(map[string]any)
			if !ok {
				child = map[string]any{}
				s[key] = child
			}
			return b.rules(child, t.Elem(), rules[i+1:])
		}
		kind := t.Kind()
		minimumKey, maximumKey := "minimum", "maximum"
		switch kind {
		case reflect.String:
			minimumKey, maximumKey = "minLength", "maxLength"
		case reflect.Array, reflect.Slice:
			minimumKey, maximumKey = "minItems", "maxItems"
		case reflect.Map:
			minimumKey, maximumKey = "minProperties", "maxProperties"
		}
		switch r.Name {
		case "required":
			s["not"] = map[string]any{"enum": []any{nil}}
			switch kind {
			case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
				setMinimum(s, minimumKey, 1)
			case reflect.Bool:
				s["not"] = map[string]any{"enum": []any{nil, false}}
			case reflect.Struct:
			default:
				s["not"] = map[string]any{"enum": []any{nil, 0}}
			}
		case "min", "gte", "max", "lte", "len":
			v, _ := schemaNumeric(r.Argument)
			if r.Name == "min" || r.Name == "gte" || r.Name == "len" {
				setMinimum(s, minimumKey, v)
			}
			if r.Name == "max" || r.Name == "lte" || r.Name == "len" {
				setMaximum(s, maximumKey, v)
			}
		case "oneof":
			values := []any{}
			for _, part := range strings.Fields(r.Argument) {
				var v any = part
				if kind != reflect.String {
					candidate := reflect.New(t)
					if err := json.Unmarshal([]byte(part), candidate.Interface()); err != nil {
						return fmt.Errorf("invalid enum %q", part)
					}
					v = candidate.Elem().Interface()
				}
				values = append(values, v)
			}
			s["enum"] = values
		case "email", "uuid", "uri":
			s["format"] = r.Name
		case "url":
			s["format"] = "uri"
		}
	}
	return nil
}
func setMinimum(s map[string]any, key string, v any) {
	if s[key] == nil || compareNumber(v, s[key]) > 0 {
		s[key] = v
	}
}
func setMaximum(s map[string]any, key string, v any) {
	if s[key] == nil || compareNumber(v, s[key]) < 0 {
		s[key] = v
	}
}
func compareNumber(a, b any) int {
	left, _ := new(big.Rat).SetString(fmt.Sprint(a))
	right, _ := new(big.Rat).SetString(fmt.Sprint(b))
	return left.Cmp(right)
}
func (b schemaWriter) operation() (map[string]any, error) {
	op := b.op
	o := map[string]any{"operationId": op.ID, "summary": op.Summary, "description": op.Description}
	if op.Tag != "" {
		o["tags"] = []string{op.Tag}
	}
	parameters := []any{}
	for _, f := range op.Types[op.Request].Fields {
		if f.Source == "json" || f.Source == "form" {
			continue
		}
		s, err := b.field(f, true)
		if err != nil {
			return nil, err
		}
		parameters = append(parameters, map[string]any{"name": f.Name, "in": f.Source, "required": f.Required, "schema": s})
	}
	if len(parameters) > 0 {
		o["parameters"] = parameters
	}
	if op.Body != "none" {
		schema, err := b.object(op.Types[op.Request], true, true)
		if err != nil {
			return nil, err
		}
		media := "application/json"
		if op.Body == "form" {
			media = "application/x-www-form-urlencoded"
		}
		required := false
		for _, f := range op.Types[op.Request].Fields {
			if (f.Source == "json" || f.Source == "form") && f.Required {
				required = true
			}
		}
		o["requestBody"] = map[string]any{"required": required, "content": map[string]any{media: map[string]any{"schema": schema}}}
	}
	response := map[string]any{"description": http.StatusText(op.Status)}
	if op.Status != 204 && op.Method != "HEAD" {
		s, err := b.ref(op.Response, false)
		if err != nil {
			return nil, err
		}
		if op.Envelope == "maltose" {
			s = map[string]any{"type": "object", "required": []string{"code", "message", "data"}, "properties": map[string]any{"code": map[string]any{"type": "integer", "enum": []any{0}}, "message": map[string]any{"type": "string"}, "data": s}}
		}
		response["content"] = map[string]any{"application/json": map[string]any{"schema": s}}
	}
	o["responses"] = map[string]any{strconv.Itoa(op.Status): response}
	return o, nil
}

// Extensions only exposes additions to business-owned document sections.
// Input maps are deep copied so retaining them cannot mutate generated content.
type Extensions struct {
	doc        map[string]any
	operations map[string]map[string]any
	statuses   map[string]int
	schemas    map[string]any
	version    string
}

func clone(value any) (any, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var out any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	err = decoder.Decode(&out)
	return out, err
}
func (e *Extensions) operation(id string) (map[string]any, error) {
	o, ok := e.operations[id]
	if !ok {
		return nil, fmt.Errorf("unknown operation_id %q", id)
	}
	return o, nil
}
func add(target map[string]any, key string, value any) error {
	if _, ok := target[key]; ok {
		return fmt.Errorf("declaration %q already exists", key)
	}
	v, err := clone(value)
	if err != nil {
		return err
	}
	target[key] = v
	return nil
}
func (e *Extensions) SecurityScheme(name string, scheme map[string]any) error {
	return add(e.doc["components"].(map[string]any)["securitySchemes"].(map[string]any), name, scheme)
}
func (e *Extensions) Security(id string, requirements []map[string][]string) error {
	o, err := e.operation(id)
	if err != nil {
		return err
	}
	return add(o, "security", requirements)
}
func (e *Extensions) Extension(id, key string, value any) error {
	if !strings.HasPrefix(key, "x-") {
		return fmt.Errorf("vendor extension must start with x-")
	}
	target := e.doc
	if id != "" {
		var err error
		target, err = e.operation(id)
		if err != nil {
			return err
		}
	}
	return add(target, key, value)
}
func (e *Extensions) ResponseHeader(id, name, description string, schema map[string]any) error {
	o, err := e.operation(id)
	if err != nil {
		return err
	}
	r := o["responses"].(map[string]any)[strconv.Itoa(e.statuses[id])].(map[string]any)
	if r["headers"] == nil {
		r["headers"] = map[string]any{}
	}
	return add(r["headers"].(map[string]any), http.CanonicalHeaderKey(name), map[string]any{"description": description, "schema": schema})
}
func (e *Extensions) ErrorResponse(id string, status int, description, media string, body reflect.Type) error {
	if status < 400 || status > 599 {
		return fmt.Errorf("error response status must be 400..599")
	}
	o, err := e.operation(id)
	if err != nil {
		return err
	}
	op := &Operation{Types: map[string]*Type{}}
	c := compiler{op: op, visiting: map[reflect.Type]bool{}}
	typeID, err := c.add(body, false, false)
	if err != nil {
		return err
	}
	b := schemaWriter{op: op, version: e.version, components: e.schemas}
	s, err := b.ref(typeID, false)
	if err != nil {
		return err
	}
	return add(o["responses"].(map[string]any), strconv.Itoa(status), map[string]any{"description": description, "content": map[string]any{media: map[string]any{"schema": s}}})
}

func canOmit(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Struct:
		return false
	case reflect.Array:
		return t.Len() == 0
	default:
		return true
	}
}

func supplement(target, extra, inferred map[string]any) error {
	for key, v := range extra {
		if key == "items" {
			item, ok := target[key].(map[string]any)
			if !ok {
				return fmt.Errorf("items schema requires an array")
			}
			rules, _ := inferred[key].(map[string]any)
			if err := supplement(item, v.(map[string]any), rules); err != nil {
				return err
			}
			continue
		}
		if existing, ok := inferred[key]; ok {
			a, _ := json.Marshal(existing)
			b, _ := json.Marshal(v)
			if string(a) != string(b) {
				return fmt.Errorf("schema %s conflicts with binding", key)
			}
		}
		switch key {
		case "minimum", "minLength", "minItems", "minProperties":
			setMinimum(target, key, v)
		case "maximum", "maxLength", "maxItems", "maxProperties":
			setMaximum(target, key, v)
		default:
			target[key] = v
		}
	}
	return nil
}

func constraintTarget(s map[string]any) map[string]any {
	if branches, ok := s["anyOf"].([]any); ok {
		return constraintTarget(branches[0].(map[string]any))
	}
	if ref, ok := s["$ref"]; ok {
		delete(s, "$ref")
		s["allOf"] = []any{map[string]any{"$ref": ref}}
	}
	return s
}

func blockYAML(node *yaml.Node) {
	node.Style = 0
	for _, child := range node.Content {
		blockYAML(child)
	}
}
