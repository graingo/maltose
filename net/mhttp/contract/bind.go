package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Presence uint8

const (
	Missing Presence = iota
	Null
	Present
)

// Input records the original request before defaults are applied. Paths include
// the source, e.g. json.profile.name or query.page.
type Input map[string]Presence

func (input Input) Presence(path string) Presence { return input[path] }

type Issue struct {
	Field  string            `json:"field"`
	Code   string            `json:"code"`
	Params map[string]string `json:"params,omitempty"`
}
type ValidationError struct {
	Issues []Issue `json:"errors"`
}

func (e *ValidationError) Error() string {
	if len(e.Issues) == 0 {
		return "request validation failed"
	}
	return e.Issues[0].Field + ": " + e.Issues[0].Code
}
func invalid(field, code, arg string) error {
	params := map[string]string{}
	if arg != "" {
		params["value"] = arg
	}
	return &ValidationError{[]Issue{{field, code, params}}}
}

// DecodeError distinguishes malformed input and unsupported media from field validation.
type DecodeError struct {
	Status int
	Cause  error
}

func (e *DecodeError) Error() string { return e.Cause.Error() }
func (e *DecodeError) Unwrap() error { return e.Cause }

var valueValidator = validator.New()

// Bind decodes each source independently and executes the compiled field rules.
// Middleware may authenticate before calling the controller handler that uses Bind.
func (op *Operation) Bind(request *http.Request, paths map[string]string, target any) (Input, error) {
	input := Input{}
	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Pointer || value.IsNil() || value.Elem().Type() != op.Types[op.Request].Go {
		return input, fmt.Errorf("binding target must be *%s", op.Types[op.Request].Go)
	}
	body := map[string]json.RawMessage{}
	form := url.Values{}
	if op.Body != "none" && request.Body != nil && request.Body != http.NoBody {
		reader := json.NewDecoder(request.Body)
		if op.Body == "json" {
			// Decode first to distinguish a truly absent body from a supplied value.
			var raw json.RawMessage
			err := reader.Decode(&raw)
			if err != nil && err != io.EOF {
				return input, &DecodeError{400, err}
			}
			if err == nil {
				media, _, _ := mime.ParseMediaType(request.Header.Get("Content-Type"))
				if media != "application/json" && !strings.HasSuffix(media, "+json") {
					return input, &DecodeError{415, fmt.Errorf("Content-Type must be application/json")}
				}
				if len(raw) == 0 || raw[0] != '{' {
					return input, &DecodeError{400, fmt.Errorf("JSON request body must be an object")}
				}
				if err = json.Unmarshal(raw, &body); err != nil {
					return input, &DecodeError{400, err}
				}
				if err = reader.Decode(&raw); err != io.EOF {
					return input, &DecodeError{400, fmt.Errorf("request body must contain one JSON value")}
				}
			}
		} else {
			media, _, _ := mime.ParseMediaType(request.Header.Get("Content-Type"))
			if media != "application/x-www-form-urlencoded" {
				return input, &DecodeError{415, fmt.Errorf("Content-Type must be application/x-www-form-urlencoded")}
			}
			if err := request.ParseForm(); err != nil {
				return input, &DecodeError{400, err}
			}
			form = request.PostForm
		}
	}
	fields := op.Types[op.Request].Fields
	known := map[string]bool{}
	for _, f := range fields {
		if f.Source == "json" {
			known[f.Name] = true
		}
	}
	if op.Unknown == "reject" {
		if err := unknownFields(body, known, "json"); err != nil {
			return input, err
		}
	}
	for _, f := range fields {
		var raw json.RawMessage
		var exists bool
		switch f.Source {
		case "json":
			raw, exists = body[f.Name]
		default:
			var values []string
			switch f.Source {
			case "path":
				v, ok := paths[f.Name]
				if ok {
					values = []string{v}
				}
			case "query":
				values = request.URL.Query()[f.Name]
			case "header":
				values = request.Header.Values(f.Name)
			case "form":
				values = form[f.Name]
			}
			exists = len(values) > 0
			if exists {
				var err error
				raw, err = parameterJSON(values, op.Types[f.Type].Go)
				if err != nil {
					return input, invalid(f.Source+"."+f.Name, "type", "")
				}
			}
		}
		if err := op.bindField(f, raw, exists, value.Elem().FieldByIndex(f.Index), f.Source+"."+f.Name, input); err != nil {
			return input, err
		}
	}
	return input, nil
}
func parameterJSON(values []string, t reflect.Type) (json.RawMessage, error) {
	t = indirect(t)
	if t.Kind() == reflect.Slice {
		items := make([]json.RawMessage, len(values))
		for i, v := range values {
			item, err := parameterJSON([]string{v}, t.Elem())
			if err != nil {
				return nil, err
			}
			items[i] = item
		}
		return json.Marshal(items)
	}
	if len(values) != 1 {
		return nil, fmt.Errorf("scalar parameter requires one value")
	}
	v := values[0]
	switch t.Kind() {
	case reflect.String:
		return json.Marshal(v)
	case reflect.Bool:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, err
		}
		return json.Marshal(b)
	default:
		if !json.Valid([]byte(v)) {
			return nil, fmt.Errorf("invalid numeric parameter")
		}
		return json.RawMessage(v), nil
	}
}
func (op *Operation) bindField(f Field, raw json.RawMessage, exists bool, target reflect.Value, path string, input Input) error {
	if !exists {
		input[path] = Missing
		if f.Required {
			return invalid(path, "required", "")
		}
		if len(f.Default) == 0 {
			return nil
		}
		raw = f.Default
	} else if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		input[path] = Null
	} else {
		input[path] = Present
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		if !f.Nullable {
			return invalid(path, "nullable", "")
		}
		if requiresValue(f.Rules) {
			return invalid(path, "required", "")
		}
		target.SetZero()
		return nil
	}
	if err := op.decode(f.Type, raw, target, path, input); err != nil {
		return err
	}
	return validateRules(target, f.Rules, path)
}
func (op *Operation) decode(id string, raw json.RawMessage, target reflect.Value, path string, input Input) error {
	n := op.Types[id]
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		if target.Kind() == reflect.Pointer {
			target.SetZero()
			return nil
		}
		return invalid(path, "type", "")
	}
	switch n.Kind {
	case "pointer":
		target.Set(reflect.New(target.Type().Elem()))
		return op.decode(n.Elem, raw, target.Elem(), path, input)
	case "object":
		var values map[string]json.RawMessage
		if err := json.Unmarshal(raw, &values); err != nil {
			return invalid(path, "type", "")
		}
		known := map[string]bool{}
		for _, f := range n.Fields {
			known[f.Name] = true
		}
		if op.Unknown == "reject" {
			if err := unknownFields(values, known, path); err != nil {
				return err
			}
		}
		for _, f := range n.Fields {
			v, exists := values[f.Name]
			if err := op.bindField(f, v, exists, target.FieldByIndex(f.Index), path+"."+f.Name, input); err != nil {
				return err
			}
		}
	case "array":
		var values []json.RawMessage
		if err := json.Unmarshal(raw, &values); err != nil {
			return invalid(path, "type", "")
		}
		if target.Kind() == reflect.Array {
			if len(values) != target.Len() {
				return invalid(path, "len", strconv.Itoa(target.Len()))
			}
		} else {
			target.Set(reflect.MakeSlice(target.Type(), len(values), len(values)))
		}
		for i, v := range values {
			child := fmt.Sprintf("%s[%d]", path, i)
			input[child] = rawPresence(v)
			if err := op.decode(n.Elem, v, target.Index(i), child, input); err != nil {
				return err
			}
		}
	case "map":
		var values map[string]json.RawMessage
		if err := json.Unmarshal(raw, &values); err != nil {
			return invalid(path, "type", "")
		}
		target.Set(reflect.MakeMap(target.Type()))
		for _, key := range sortedKeys(values) {
			v := reflect.New(target.Type().Elem()).Elem()
			child := path + "[" + strconv.Quote(key) + "]"
			input[child] = rawPresence(values[key])
			if err := op.decode(n.Elem, values[key], v, child, input); err != nil {
				return err
			}
			target.SetMapIndex(reflect.ValueOf(key).Convert(target.Type().Key()), v)
		}
	default:
		if err := json.Unmarshal(raw, target.Addr().Interface()); err != nil {
			return invalid(path, "type", "")
		}
	}
	return nil
}
func rawPresence(raw json.RawMessage) Presence {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return Null
	}
	return Present
}
func unknownFields(values map[string]json.RawMessage, known map[string]bool, path string) error {
	for _, key := range sortedKeys(values) {
		if !known[key] {
			return invalid(path+"."+key, "unknown", "")
		}
	}
	return nil
}
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
func validateRules(v reflect.Value, rules []Rule, path string) error {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			if requiresValue(rules) {
				return invalid(path, "required", "")
			}
			return nil
		}
		v = v.Elem()
	}
	for i, r := range rules {
		if r.Name == "dive" {
			if v.Kind() == reflect.Map {
				keys := v.MapKeys()
				sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
				for _, k := range keys {
					if err := validateRules(v.MapIndex(k), rules[i+1:], path+"["+strconv.Quote(k.String())+"]"); err != nil {
						return err
					}
				}
			} else {
				for j := 0; j < v.Len(); j++ {
					if err := validateRules(v.Index(j), rules[i+1:], fmt.Sprintf("%s[%d]", path, j)); err != nil {
						return err
					}
				}
			}
			return nil
		}
		if r.Name == "required" {
			zero := v.IsZero()
			if v.Kind() == reflect.Slice || v.Kind() == reflect.Map || v.Kind() == reflect.Array {
				zero = v.Len() == 0
			}
			if zero {
				return invalid(path, r.Name, "")
			}
			continue
		}
		if r.Name == "oneof" {
			matched := false
			for _, part := range strings.Fields(r.Argument) {
				matches := v.Kind() == reflect.String && v.String() == part
				if v.Kind() != reflect.String {
					candidate := reflect.New(v.Type())
					if err := json.Unmarshal([]byte(part), candidate.Interface()); err == nil {
						matches = reflect.DeepEqual(v.Interface(), candidate.Elem().Interface())
					}
				}
				if matches {
					matched = true
					break
				}
			}
			if !matched {
				return invalid(path, r.Name, r.Argument)
			}
			continue
		}
		tag := r.Name
		if r.Argument != "" {
			tag += "=" + r.Argument
		}
		if err := valueValidator.Var(v.Interface(), tag); err != nil {
			return invalid(path, r.Name, r.Argument)
		}
	}
	return nil
}
