package contract_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/graingo/maltose/net/mhttp/contract"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func metadataDocument(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var doc map[string]any
	require.NoError(t, yaml.Unmarshal(data, &doc))
	return doc
}

func TestDocumentMetadata(t *testing.T) {
	op := operation(t)
	for _, version := range []string{"3.0.0", "3.1.0"} {
		for _, format := range []string{"json", "yaml"} {
			t.Run(version+format, func(t *testing.T) {
				options := contract.Options{Version: version, Format: format}
				base, err := contract.Generate([]*contract.Operation{op}, options, nil)
				require.NoError(t, err)
				require.Equal(t, map[string]any{"title": "API", "version": "1.0.0"}, metadataDocument(t, base.Document)["info"])
				configure := func(e *contract.Extensions) error {
					contact := &contract.Contact{Name: "maintainer", URL: "https://example.com", Email: "api@example.com"}
					license := &contract.License{Name: "MIT", URL: "https://example.com/license"}
					if err := e.Info(contract.Info{Title: "Example API", Version: "0.1.0"}); err != nil {
						return err
					}
					if err := e.Info(contract.Info{Title: "Example API", Description: "Description", TermsOfService: "/terms", Contact: contact, License: license}); err != nil {
						return err
					}
					contact.Name = "mutated"
					license.Name = "mutated"
					server := contract.Server{URL: "https://example.com/api", Description: "Production"}
					if err := e.Server(server); err != nil {
						return err
					}
					server.URL = "https://mutated.example.com"
					if err := e.Server(contract.Server{URL: "/", Description: "Current server"}); err != nil {
						return err
					}
					docs := &contract.ExternalDocs{URL: "https://example.com/docs", Description: "Manual"}
					if err := e.Tag(contract.Tag{Name: "Z", Description: "First", ExternalDocs: docs}); err != nil {
						return err
					}
					if err := e.Tag(contract.Tag{Name: "A", Description: "Second"}); err != nil {
						return err
					}
					if err := e.ExternalDocs(*docs); err != nil {
						return err
					}
					docs.URL = "mutated"
					return nil
				}
				artifact, err := contract.Generate([]*contract.Operation{op}, options, configure)
				require.NoError(t, err)
				again, err := contract.Generate([]*contract.Operation{op}, options, configure)
				require.NoError(t, err)
				require.Equal(t, artifact, again)
				doc := metadataDocument(t, artifact.Document)
				info := doc["info"].(map[string]any)
				require.Equal(t, "Example API", info["title"])
				require.Equal(t, "0.1.0", info["version"])
				require.Equal(t, "Description", info["description"])
				require.Equal(t, "maintainer", info["contact"].(map[string]any)["name"])
				require.Equal(t, "MIT", info["license"].(map[string]any)["name"])
				tags := doc["tags"].([]any)
				require.Equal(t, "Z", tags[0].(map[string]any)["name"])
				require.Equal(t, "A", tags[1].(map[string]any)["name"])
				require.Equal(t, "https://example.com/docs", tags[0].(map[string]any)["externalDocs"].(map[string]any)["url"])
				servers := doc["servers"].([]any)
				require.Equal(t, "/", servers[1].(map[string]any)["url"])
				require.Equal(t, "https://example.com/api", servers[0].(map[string]any)["url"])
				require.NoError(t, contract.CheckArtifact(artifact.Document, artifact.Manifest, []*contract.Operation{op}))
				var before, after contract.Manifest
				require.NoError(t, json.Unmarshal(base.Manifest, &before))
				require.NoError(t, json.Unmarshal(artifact.Manifest, &after))
				require.Equal(t, before.Operations, after.Operations)
				require.NotEqual(t, before.DocumentSHA256, after.DocumentSHA256)
			})
		}
	}
}

func TestMetadataErrors(t *testing.T) {
	cases := []struct {
		name      string
		configure func(*contract.Extensions) error
		want      string
	}{
		{"conflict", func(e *contract.Extensions) error {
			if err := e.Info(contract.Info{Title: "A"}); err != nil {
				return err
			}
			return e.Info(contract.Info{Title: "B"})
		}, "info.title conflicts"},
		{"blank title", func(e *contract.Extensions) error { return e.Info(contract.Info{Title: " "}) }, "info.title"},
		{"blank version", func(e *contract.Extensions) error { return e.Info(contract.Info{Version: " "}) }, "info.version"},
		{"email", func(e *contract.Extensions) error {
			return e.Info(contract.Info{Contact: &contract.Contact{Email: "bad"}})
		}, "email"},
		{"contact url", func(e *contract.Extensions) error {
			return e.Info(contract.Info{Contact: &contract.Contact{URL: "https://["}})
		}, "contact.url"},
		{"license name", func(e *contract.Extensions) error { return e.Info(contract.Info{License: &contract.License{}}) }, "license.name"},
		{"license url", func(e *contract.Extensions) error {
			return e.Info(contract.Info{License: &contract.License{Name: "MIT", URL: "%xx"}})
		}, "license.url"},
		{"terms", func(e *contract.Extensions) error { return e.Info(contract.Info{TermsOfService: "bad url"}) }, "termsOfService"},
		{"server required", func(e *contract.Extensions) error { return e.Server(contract.Server{}) }, "server.url"},
		{"server malformed", func(e *contract.Extensions) error { return e.Server(contract.Server{URL: "https://["}) }, "server.url"},
		{"server duplicate", func(e *contract.Extensions) error {
			if err := e.Server(contract.Server{URL: "/"}); err != nil {
				return err
			}
			return e.Server(contract.Server{URL: "/"})
		}, "already exists"},
		{"server variable", func(e *contract.Extensions) error {
			return e.Server(contract.Server{URL: "https://{host}"})
		}, "server.url"},
		{"tag required", func(e *contract.Extensions) error { return e.Tag(contract.Tag{}) }, "tag.name"},
		{"tag duplicate", func(e *contract.Extensions) error {
			if err := e.Tag(contract.Tag{Name: "A"}); err != nil {
				return err
			}
			return e.Tag(contract.Tag{Name: "A"})
		}, "already exists"},
		{"tag docs", func(e *contract.Extensions) error {
			return e.Tag(contract.Tag{Name: "A", ExternalDocs: &contract.ExternalDocs{}})
		}, "externalDocs.url"},
		{"docs required", func(e *contract.Extensions) error { return e.ExternalDocs(contract.ExternalDocs{}) }, "externalDocs.url"},
		{"docs duplicate", func(e *contract.Extensions) error {
			if err := e.ExternalDocs(contract.ExternalDocs{URL: "/docs"}); err != nil {
				return err
			}
			return e.ExternalDocs(contract.ExternalDocs{URL: "/docs"})
		}, "already exists"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			artifact, err := contract.Generate([]*contract.Operation{operation(t)}, contract.Options{}, tc.configure)
			require.ErrorContains(t, err, tc.want)
			require.Empty(t, artifact)
		})
	}
}

func TestMetadataInfoAtomicMerge(t *testing.T) {
	artifact, err := contract.Generate([]*contract.Operation{operation(t)}, contract.Options{}, func(e *contract.Extensions) error {
		if err := e.Info(contract.Info{Title: "Original"}); err != nil {
			return err
		}
		if err := e.Info(contract.Info{Title: "Conflict", Description: "Must stay absent"}); err == nil {
			return fmt.Errorf("expected conflict")
		}
		return e.Info(contract.Info{Contact: &contract.Contact{Name: "Owner"}})
	})
	require.NoError(t, err)
	info := metadataDocument(t, artifact.Document)["info"].(map[string]any)
	require.Equal(t, "Original", info["title"])
	require.Equal(t, "1.0.0", info["version"])
	require.NotContains(t, info, "description")
}
