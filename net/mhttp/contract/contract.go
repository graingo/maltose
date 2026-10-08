// Package contract compiles Maltose DTOs into a shared HTTP contract.
package contract

import (
	"crypto/sha256"
	"encoding"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/graingo/maltose/util/mmeta"
)

const Version = 1

type Rule struct {
	Name     string
	Argument string
}
type Field struct {
	Name        string
	Source      string
	Type        string
	Required    bool
	Nullable    bool
	Rules       []Rule
	Default     json.RawMessage `json:",omitempty"`
	Description string          `json:",omitempty"`
	Schema      map[string]any  `json:",omitempty"`
	OmitEmpty   bool            `json:",omitempty"`
	Index       []int
}
type Type struct {
	ID     string
	Kind   string
	Elem   string       `json:",omitempty"`
	Length int          `json:",omitempty"`
	Fields []Field      `json:",omitempty"`
	Go     reflect.Type `json:"-"`
}
type Operation struct {
	Method       string
	Path         string
	RelativePath string
	Group        string
	ID           string
	Summary      string
	Description  string
	Tag          string
	Status       int
	Body         string
	Envelope     string
	Unknown      string
	Request      string
	Response     string
	Types        map[string]*Type
}

func TypeOf[T any]() reflect.Type { return reflect.TypeOf((*T)(nil)).Elem() }

// Compile reads one Req/Res pair. The returned graph is immutable after compilation.
func Compile(req, res reflect.Type) (*Operation, error) {
	req, res = indirect(req), indirect(res)
	if req.Kind() != reflect.Struct || res.Kind() != reflect.Struct || req.PkgPath() != res.PkgPath() || !strings.HasSuffix(req.Name(), "Req") || res.Name() != strings.TrimSuffix(req.Name(), "Req")+"Res" {
		return nil, fmt.Errorf("contract requires a named XxxReq/XxxRes struct pair")
	}
	var meta reflect.StructTag
	for i := 0; i < req.NumField(); i++ {
		f := req.Field(i)
		if f.Anonymous && f.Type == reflect.TypeOf(mmeta.Meta{}) {
			meta = f.Tag
		}
	}
	if meta == "" {
		return nil, fmt.Errorf("%s requires m.Meta", req)
	}
	op := &Operation{Method: strings.ToUpper(meta.Get("method")), RelativePath: meta.Get("path"), Group: meta.Get("group"), ID: meta.Get("operation_id"), Summary: meta.Get("summary"), Description: meta.Get("dc"), Tag: meta.Get("tag"), Status: 200, Envelope: meta.Get("envelope"), Body: meta.Get("body"), Unknown: meta.Get("unknown"), Types: map[string]*Type{}}
	switch op.Method {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD":
	default:
		return nil, fmt.Errorf("unsupported method %q", op.Method)
	}
	if !strings.HasPrefix(op.RelativePath, "/") || strings.ContainsAny(op.RelativePath+op.Group, "{}*?#") {
		return nil, fmt.Errorf("path must start with / and use :name parameters")
	}
	if op.Group != "" && !strings.HasPrefix(op.Group, "/") {
		return nil, fmt.Errorf("group must start with /")
	}
	if op.Group == "/" {
		op.Group = ""
	}
	op.Path = path.Join("/", op.Group, op.RelativePath)
	if op.Path != "/" && strings.HasSuffix(op.RelativePath, "/") {
		op.Path += "/"
	}
	if op.ID == "" {
		op.ID = req.PkgPath() + "." + strings.TrimSuffix(req.Name(), "Req")
	}
	if raw := meta.Get("status"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid status %q", raw)
		}
		op.Status = n
	}
	switch op.Status {
	case 200, 201, 202, 204:
	default:
		return nil, fmt.Errorf("unsupported success status %d", op.Status)
	}
	if op.Envelope == "" {
		op.Envelope = "none"
	}
	if op.Envelope != "none" && op.Envelope != "maltose" {
		return nil, fmt.Errorf("envelope must be none or maltose")
	}
	if op.Unknown == "" {
		op.Unknown = "reject"
	}
	if op.Unknown != "reject" && op.Unknown != "allow" {
		return nil, fmt.Errorf("unknown must be reject or allow")
	}
	c := compiler{op: op, visiting: map[reflect.Type]bool{}}
	var err error
	op.Request, err = c.add(req, true, true)
	if err != nil {
		return nil, err
	}
	op.Response, err = c.add(res, false, false)
	if err != nil {
		return nil, err
	}
	sources := map[string]bool{}
	params := map[string]bool{}
	for _, f := range op.Types[op.Request].Fields {
		sources[f.Source] = true
		if f.Source == "path" {
			params[f.Name] = true
		}
	}
	if sources["json"] && sources["form"] {
		return nil, fmt.Errorf("request uses both JSON and form bodies")
	}
	inferred := "none"
	if sources["json"] {
		inferred = "json"
	}
	if sources["form"] {
		inferred = "form"
	}
	if op.Body == "" {
		op.Body = inferred
	}
	if op.Body != inferred {
		return nil, fmt.Errorf("body %q conflicts with field sources %q", op.Body, inferred)
	}
	for _, segment := range strings.Split(op.Path, "/") {
		if strings.HasPrefix(segment, ":") {
			name := segment[1:]
			if !params[name] {
				return nil, fmt.Errorf("path parameter %q requires a path field", name)
			}
			delete(params, name)
		}
	}
	if len(params) > 0 {
		return nil, fmt.Errorf("path fields must match route parameters")
	}
	for _, n := range op.Types {
		for _, f := range n.Fields {
			if len(f.Default) > 0 {
				target := reflect.New(op.Types[f.Type].Go).Elem()
				if err := op.bindField(f, nil, false, target, f.Source+"."+f.Name, Input{}); err != nil {
					return nil, fmt.Errorf("invalid default: %w", err)
				}
			}
		}
	}
	return op, nil
}
func indirect(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}
func typeID(t reflect.Type, request, root bool, unknown string) string {
	identity := typeIdentity(t) + ":" + strconv.FormatBool(request) + ":" + strconv.FormatBool(root)
	if request {
		identity += ":" + unknown
	}
	sum := sha256.Sum256([]byte(identity))
	name := strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' {
			return r
		}
		return '_'
	}, indirect(t).Name())
	if name == "" {
		name = "Value"
	}
	return name + "_" + hex.EncodeToString(sum[:])
}

type compiler struct {
	op       *Operation
	visiting map[reflect.Type]bool
}

func (c *compiler) add(t reflect.Type, request, root bool) (string, error) {
	id := typeID(t, request, root, c.op.Unknown)
	if _, ok := c.op.Types[id]; ok {
		return id, nil
	}
	n := &Type{ID: id, Go: t}
	c.op.Types[id] = n
	if t == reflect.TypeOf(time.Time{}) {
		n.Kind = "time"
		return id, nil
	}
	if t.Kind() != reflect.Pointer && hasCustomCodec(t) {
		return "", fmt.Errorf("%s: custom JSON codecs require an explicit adapter", t)
	}
	switch t.Kind() {
	case reflect.Interface:
		if request || t.NumMethod() != 0 {
			return "", fmt.Errorf("interface fields are supported only as free-form JSON response values")
		}
		n.Kind = "any"
	case reflect.Pointer:
		n.Kind = "pointer"
	case reflect.String:
		n.Kind = "string"
	case reflect.Bool:
		n.Kind = "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n.Kind = "integer"
	case reflect.Float32, reflect.Float64:
		n.Kind = "number"
	case reflect.Slice, reflect.Array:
		n.Kind = "array"
		if t.Kind() == reflect.Array {
			n.Length = t.Len()
		}
		if t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 {
			return "", fmt.Errorf("byte slices use custom JSON encoding; declare a base64 string")
		}
	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			return "", fmt.Errorf("%s: map keys must be strings", t)
		}
		n.Kind = "map"
	case reflect.Struct:
		n.Kind = "object"
	default:
		return "", fmt.Errorf("unsupported contract type %s", t)
	}
	switch n.Kind {
	case "pointer", "array", "map":
		elem, err := c.add(t.Elem(), request, false)
		if err != nil {
			return "", err
		}
		n.Elem = elem
	case "object":
		fields, err := c.fields(t, request, root, nil)
		if err != nil {
			return "", err
		}
		n.Fields = fields
	}
	return id, nil
}
func (c *compiler) fields(t reflect.Type, request, root bool, prefix []int) ([]Field, error) {
	if c.visiting[t] {
		return nil, fmt.Errorf("recursive inline %s", t)
	}
	c.visiting[t] = true
	defer delete(c.visiting, t)
	var fields []Field
	seen := map[string]bool{}
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.Anonymous && !sf.IsExported() {
			return nil, fmt.Errorf("inline DTO type %s must be exported", sf.Type)
		}
		if sf.Type == reflect.TypeOf(mmeta.Meta{}) || !sf.IsExported() {
			continue
		}
		index := append(append([]int{}, prefix...), i)
		opts := map[string]bool{}
		for _, s := range strings.Split(sf.Tag.Get("field"), ",") {
			if s != "" {
				if s != "required" && s != "nullable" && s != "inline" {
					return nil, fmt.Errorf("%s.%s: unknown field option %q", t, sf.Name, s)
				}
				opts[s] = true
			}
		}
		if sf.Anonymous {
			if !opts["inline"] || sf.Type.Kind() != reflect.Struct || len(opts) != 1 || sf.Tag.Get("binding") != "" {
				return nil, fmt.Errorf("%s.%s: anonymous value struct requires field:\"inline\" and no group constraints", t, sf.Name)
			}
			for _, key := range []string{"json", "query", "path", "header", "form", "schema", "default"} {
				if sf.Tag.Get(key) != "" {
					return nil, fmt.Errorf("inline field %s only declares field:\"inline\"", sf.Name)
				}
			}
			nested, err := c.fields(sf.Type, request, root, index)
			if err != nil {
				return nil, err
			}
			fields = append(fields, nested...)
			continue
		}
		if opts["inline"] {
			return nil, fmt.Errorf("inline requires an anonymous value struct")
		}
		f := Field{Index: index, Required: opts["required"], Nullable: opts["nullable"], Description: sf.Tag.Get("dc")}
		sources := []string{"json"}
		if request && !root {
			for _, source := range []string{"path", "query", "header", "form"} {
				if sf.Tag.Get(source) != "" {
					return nil, fmt.Errorf("nested field %s uses JSON source", sf.Name)
				}
			}
		}
		if request && root {
			sources = []string{"path", "query", "header", "form", "json"}
		}
		for _, source := range sources {
			raw, ok := sf.Tag.Lookup(source)
			if !ok || raw == "-" {
				continue
			}
			parts := strings.Split(raw, ",")
			if parts[0] == "" {
				return nil, fmt.Errorf("%s.%s: explicit field name is required", t, sf.Name)
			}
			if f.Source != "" {
				return nil, fmt.Errorf("%s.%s: use one field source", t, sf.Name)
			}
			f.Source = source
			f.Name = parts[0]
			for _, p := range parts[1:] {
				if p != "omitempty" {
					return nil, fmt.Errorf("%s.%s: unsupported tag option %q", t, sf.Name, p)
				}
				f.OmitEmpty = true
			}
		}
		if f.Source == "" {
			if sf.Tag.Get("json") == "-" {
				continue
			}
			return nil, fmt.Errorf("%s.%s: explicit source tag is required", t, sf.Name)
		}
		if f.Source == "path" {
			f.Required = true
		}
		if f.Source == "header" {
			f.Name = http.CanonicalHeaderKey(f.Name)
		}
		id, err := c.add(sf.Type, request, false)
		if err != nil {
			return nil, err
		}
		f.Type = id
		if request {
			f.Rules, err = parseRules(sf.Tag.Get("binding"), sf.Type)
			if err != nil {
				return nil, fmt.Errorf("%s.%s: %w", t, sf.Name, err)
			}
			if requiresValue(f.Rules) {
				if f.Nullable {
					return nil, fmt.Errorf("binding required rejects null; use field:required,nullable for presence-only validation")
				}
				f.Required = true
			}
			if f.Nullable && (sf.Type.Kind() != reflect.Pointer && sf.Type.Kind() != reflect.Map && sf.Type.Kind() != reflect.Slice) {
				return nil, fmt.Errorf("nullable requires a pointer, slice or map: %s", sf.Name)
			}
			if f.Nullable && f.Source != "json" {
				return nil, fmt.Errorf("nullable applies to JSON fields")
			}
			if raw, ok := sf.Tag.Lookup("default"); ok {
				base := indirect(sf.Type)
				if base.Kind() == reflect.Struct || base.Kind() == reflect.Array || base.Kind() == reflect.Slice || base.Kind() == reflect.Map {
					return nil, fmt.Errorf("default requires a scalar field")
				}
				if f.Required {
					return nil, fmt.Errorf("required field %s cannot have a default", sf.Name)
				}
				if err := json.Unmarshal([]byte(raw), reflect.New(sf.Type).Interface()); err != nil {
					return nil, fmt.Errorf("invalid default for %s: %w", sf.Name, err)
				}
				f.Default = json.RawMessage(raw)
			}
			if f.Source != "json" {
				base := indirect(sf.Type)
				if base.Kind() == reflect.Slice {
					if f.Source == "header" || f.Source == "path" {
						return nil, fmt.Errorf("%s parameters require scalar fields", f.Source)
					}
					base = indirect(base.Elem())
				}
				switch base.Kind() {
				case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
				default:
					return nil, fmt.Errorf("%s parameters require scalars or scalar slices", f.Source)
				}
			}
		}
		f.Schema, err = parseSchema(sf.Tag.Get("schema"), sf.Type)
		if err != nil {
			return nil, fmt.Errorf("%s.%s: %w", t, sf.Name, err)
		}
		fields = append(fields, f)
	}
	sort.Slice(fields, func(i, j int) bool { a, b := fields[i], fields[j]; return a.Source+":"+a.Name < b.Source+":"+b.Name })
	for _, f := range fields {
		key := f.Source + ":" + f.Name
		if seen[key] {
			return nil, fmt.Errorf("duplicate field %s in %s", key, t)
		}
		seen[key] = true
	}
	return fields, nil
}

// requiresValue checks this value's rules; dive starts the element's scope.
func requiresValue(rules []Rule) bool {
	for _, rule := range rules {
		switch rule.Name {
		case "required":
			return true
		case "dive":
			return false
		}
	}
	return false
}

func parseRules(raw string, t reflect.Type) ([]Rule, error) {
	var out []Rule
	for _, part := range strings.Split(raw, ",") {
		if part == "" {
			continue
		}
		name, arg, _ := strings.Cut(part, "=")
		r := Rule{name, arg}
		base := indirect(t)
		switch name {
		case "dive":
			if base.Kind() != reflect.Slice && base.Kind() != reflect.Array && base.Kind() != reflect.Map {
				return nil, fmt.Errorf("dive requires a collection")
			}
			t = base.Elem()
		case "required":
			if base.Kind() == reflect.Struct {
				return nil, fmt.Errorf("object presence uses field:required")
			}
		case "min", "max", "gte", "lte", "len":
			switch base.Kind() {
			case reflect.String, reflect.Array, reflect.Slice, reflect.Map, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
			default:
				return nil, fmt.Errorf("%s requires a number, string or collection", name)
			}
			if err := validateBound(base, arg); err != nil {
				return nil, fmt.Errorf("invalid %s argument %q", name, arg)
			}
		case "oneof":
			if base.Kind() != reflect.String && base.Kind() != reflect.Bool && !(base.Kind() >= reflect.Int && base.Kind() <= reflect.Float64) {
				return nil, fmt.Errorf("oneof requires scalar values")
			}
			if strings.ContainsAny(arg, "'|,") {
				return nil, fmt.Errorf("oneof uses space-separated scalar values")
			}
			if len(strings.Fields(arg)) == 0 {
				return nil, fmt.Errorf("oneof requires values")
			}
			for _, v := range strings.Fields(arg) {
				if base.Kind() != reflect.String {
					if err := json.Unmarshal([]byte(v), reflect.New(base).Interface()); err != nil {
						return nil, fmt.Errorf("invalid oneof value %q", v)
					}
				}
			}
		case "email", "url", "uri", "uuid":
			if base.Kind() != reflect.String {
				return nil, fmt.Errorf("%s requires a string", name)
			}
		default:
			return nil, fmt.Errorf("unsupported binding rule %q", name)
		}
		if (name == "required" || name == "dive" || name == "email" || name == "url" || name == "uri" || name == "uuid") && arg != "" {
			return nil, fmt.Errorf("%s accepts no argument", name)
		}
		out = append(out, r)
	}
	return out, nil
}
func parseSchema(raw string, t reflect.Type) (map[string]any, error) {
	t = indirect(t)
	out := map[string]any{}
	for _, p := range strings.Split(raw, ",") {
		if p == "" {
			continue
		}
		key, value, ok := strings.Cut(p, "=")
		if !ok {
			return nil, fmt.Errorf("schema requires key=value")
		}
		if strings.HasPrefix(key, "items.") {
			if t.Kind() != reflect.Slice && t.Kind() != reflect.Array {
				return nil, fmt.Errorf("items schema requires an array")
			}
			child, err := parseSchema(strings.TrimPrefix(p, "items."), t.Elem())
			if err != nil {
				return nil, err
			}
			if out["items"] == nil {
				out["items"] = map[string]any{}
			}
			for k, v := range child {
				out["items"].(map[string]any)[k] = v
			}
			continue
		}
		switch key {
		case "format", "pattern", "title":
			out[key] = value
		case "enum":
			values := []any{}
			for _, v := range strings.Split(value, "|") {
				if t.Kind() == reflect.Array || t.Kind() == reflect.Slice || t.Kind() == reflect.Map || t.Kind() == reflect.Struct {
					return nil, fmt.Errorf("enum requires a scalar; use items.enum for array elements")
				}
				var value any = v
				if t.Kind() != reflect.String {
					candidate := reflect.New(t)
					if err := json.Unmarshal([]byte(v), candidate.Interface()); err != nil {
						return nil, fmt.Errorf("invalid schema enum %q", v)
					}
					value = candidate.Elem().Interface()
				}
				values = append(values, value)
			}
			out[key] = values
		case "minimum", "maximum", "minLength", "maxLength", "minItems", "maxItems", "minProperties", "maxProperties":
			v, err := schemaNumeric(value)
			if err != nil {
				return nil, err
			}
			out[key] = v
		case "readOnly", "writeOnly", "deprecated":
			v, err := strconv.ParseBool(value)
			if err != nil {
				return nil, err
			}
			out[key] = v
		case "default", "example":
			var v any
			if err := json.Unmarshal([]byte(value), &v); err != nil {
				v = value
			}
			out[key] = v
		default:
			return nil, fmt.Errorf("unsupported schema key %q", key)
		}
	}
	return out, nil
}

// Fingerprint covers normalized HTTP metadata, both DTO graphs, and constraints.
func (op *Operation) Fingerprint() string {
	b, _ := json.Marshal(op)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
func (op *Operation) OpenAPIPath() string {
	parts := strings.Split(op.Path, "/")
	for i, s := range parts {
		if strings.HasPrefix(s, ":") {
			parts[i] = "{" + s[1:] + "}"
		}
	}
	return strings.Join(parts, "/")
}

func hasCustomCodec(t reflect.Type) bool {
	for _, codec := range []reflect.Type{reflect.TypeOf((*json.Marshaler)(nil)).Elem(), reflect.TypeOf((*json.Unmarshaler)(nil)).Elem(), reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem(), reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()} {
		if t.Implements(codec) || reflect.PointerTo(t).Implements(codec) {
			return true
		}
	}
	return false
}
func validateBound(t reflect.Type, arg string) error {
	switch t.Kind() {
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(arg, t.Bits())
		if err != nil {
			return err
		}
		if math.IsInf(v, 0) || math.IsNaN(v) {
			return fmt.Errorf("bound must be finite")
		}
		return nil
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		value, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			return err
		}
		if value < 0 {
			return fmt.Errorf("length bound must be non-negative")
		}
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		_, err := strconv.ParseInt(arg, 10, 64)
		return err
	default:
		_, err := strconv.ParseUint(arg, 10, 64)
		return err
	}
}

// ValidateOperations checks the route and operation identities as a set.
func ValidateOperations(operations []*Operation) error {
	ids, paths := map[string]bool{}, map[string]bool{}
	for _, op := range operations {
		key := op.Method + " " + op.Path
		if ids[op.ID] {
			return fmt.Errorf("duplicate operation_id %q", op.ID)
		}
		if paths[key] {
			return fmt.Errorf("duplicate route %s", key)
		}
		ids[op.ID] = true
		paths[key] = true
	}
	return nil
}

func typeIdentity(t reflect.Type) string {
	if t.Name() != "" {
		return t.PkgPath() + "." + t.Name()
	}
	switch t.Kind() {
	case reflect.Pointer:
		return "*" + typeIdentity(t.Elem())
	case reflect.Slice:
		return "[]" + typeIdentity(t.Elem())
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), typeIdentity(t.Elem()))
	case reflect.Map:
		return "map[" + typeIdentity(t.Key()) + "]" + typeIdentity(t.Elem())
	case reflect.Struct:
		var b strings.Builder
		b.WriteString("struct{")
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			fmt.Fprintf(&b, "%q:%q:%t:%s:%q;", f.PkgPath, f.Name, f.Anonymous, typeIdentity(f.Type), f.Tag)
		}
		return b.String() + "}"
	default:
		return t.String()
	}
}

func schemaNumeric(raw string) (any, error) {
	if value, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return value, nil
	}
	if value, err := strconv.ParseUint(raw, 10, 64); err == nil {
		return value, nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil, err
	}
	if math.IsInf(value, 0) || math.IsNaN(value) {
		return nil, fmt.Errorf("schema number must be finite")
	}
	return value, nil
}
