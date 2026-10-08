package contract

import (
	"fmt"
	"net/mail"
	"net/url"
	"reflect"
	"sort"
	"strings"
)

// Info describes the API document. Empty fields leave existing declarations intact.
// Title and Version default to API and 1.0.0 after configuration completes.
type Info struct {
	Title          string   `json:"title,omitempty"`
	Version        string   `json:"version,omitempty"`
	Description    string   `json:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
}

type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// License uses the fields shared by OpenAPI 3.0 and 3.1.
type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Server describes a fixed absolute or relative server URL.
type Server struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

type Tag struct {
	Name         string        `json:"name"`
	Description  string        `json:"description,omitempty"`
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`
}

type ExternalDocs struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// Info merges explicit fields. Identical declarations are accepted; conflicting
// declarations fail. Contact and License each form a single declaration.
func (e *Extensions) Info(info Info) error {
	for _, field := range []struct{ name, value string }{
		{"info.title", info.Title}, {"info.version", info.Version},
	} {
		if field.value != "" && strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("%s must contain text", field.name)
		}
	}
	if err := metadataURL("info.termsOfService", info.TermsOfService, false); err != nil {
		return err
	}
	if info.Contact != nil {
		if err := metadataURL("info.contact.url", info.Contact.URL, false); err != nil {
			return err
		}
		if email := info.Contact.Email; email != "" {
			address, err := mail.ParseAddress(email)
			if err != nil || address.Address != email {
				return fmt.Errorf("info.contact.email must be an email address")
			}
		}
	}
	if info.License != nil {
		if strings.TrimSpace(info.License.Name) == "" {
			return fmt.Errorf("info.license.name is required")
		}
		if err := metadataURL("info.license.url", info.License.URL, false); err != nil {
			return err
		}
	}
	value, err := clone(info)
	if err != nil {
		return err
	}
	fields := value.(map[string]any)
	target := e.doc["info"].(map[string]any)
	// Check every field before applying the declaration. Sorted keys give stable errors.
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if existing, ok := target[key]; ok && !reflect.DeepEqual(existing, fields[key]) {
			return fmt.Errorf("info.%s conflicts with an existing declaration", key)
		}
	}
	for _, key := range keys {
		target[key] = fields[key]
	}
	return nil
}

// Server appends a server in display order. URLs uniquely identify declarations.
func (e *Extensions) Server(server Server) error {
	if err := metadataURL("server.url", server.URL, true); err != nil {
		return err
	}
	return e.appendMetadata("servers", "url", server.URL, server)
}

// Tag adds top-level documentation; operation tag membership remains in m.Meta.
func (e *Extensions) Tag(tag Tag) error {
	if strings.TrimSpace(tag.Name) == "" {
		return fmt.Errorf("tag.name is required")
	}
	if tag.ExternalDocs != nil {
		if err := metadataURL("tag.externalDocs.url", tag.ExternalDocs.URL, true); err != nil {
			return err
		}
	}
	return e.appendMetadata("tags", "name", tag.Name, tag)
}

func (e *Extensions) ExternalDocs(docs ExternalDocs) error {
	if err := metadataURL("externalDocs.url", docs.URL, true); err != nil {
		return err
	}
	return add(e.doc, "externalDocs", docs)
}

func (e *Extensions) appendMetadata(section, key, identity string, value any) error {
	entries, _ := e.doc[section].([]any)
	for _, entry := range entries {
		if entry.(map[string]any)[key] == identity {
			return fmt.Errorf("%s declaration %q already exists", section, identity)
		}
	}
	copied, err := clone(value)
	if err != nil {
		return err
	}
	e.doc[section] = append(entries, copied)
	return nil
}

func metadataURL(field, value string, required bool) error {
	if value == "" {
		if required {
			return fmt.Errorf("%s is required", field)
		}
		return nil
	}
	if strings.ContainsAny(value, " \t\r\n{}") {
		return fmt.Errorf("%s must be a valid URL reference", field)
	}
	if _, err := url.Parse(value); err != nil {
		return fmt.Errorf("%s: %w", field, err)
	}
	return nil
}
