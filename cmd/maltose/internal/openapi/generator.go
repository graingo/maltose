// Package openapi handles the generation of the OpenAPI specification.
package openapi

import (
	"bytes"
	"encoding/json"
	"os"
	"path"
	"path/filepath"

	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/graingo/maltose/errors/merror"
	"gopkg.in/yaml.v3"
)

// Generate creates the final OpenAPI specification file.
func Generate(src, outputFile, format string) error {
	utils.PrintInfo("🔍 Scanning directory: {{.Path}}", utils.TplData{"Path": filepath.Base(src)})

	// Step 1: Parse the source code in the directory.
	// The parser will return a structured representation of the API definitions.
	apiDefs, allStructs, err := ParseDir(src)
	if err != nil {
		return merror.Wrap(err, "failed to parse source directory")
	}

	if len(apiDefs) == 0 {
		return merror.Newf("no API definitions (structs with m.Meta) found in %s", src)
	}

	utils.PrintInfo("ℹ️  Found {{.Count}} API endpoint definitions.", utils.TplData{"Count": len(apiDefs)})

	moduleName, _, err := utils.GetModuleInfo(".")
	if err != nil {
		return merror.Wrap(err, "failed to get project name")
	}
	projectName := path.Base(moduleName)

	// Step 2: Build the OpenAPI specification from the parsed definitions.
	spec, err := BuildSpec(apiDefs, projectName, allStructs)
	if err != nil {
		return merror.Wrap(err, "failed to build OpenAPI spec")
	}

	var outputBytes []byte
	var marshalErr error

	if format == "json" {
		// Step 3 (JSON): Marshal the specification to JSON.
		outputBytes, marshalErr = spec.MarshalJSON()
		if marshalErr == nil {
			// Pretty-print the JSON
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, outputBytes, "", "  "); err == nil {
				outputBytes = prettyJSON.Bytes()
			}
		}
	} else {
		// Step 3 (YAML): Marshal the specification to YAML, manually controlling field order.
		var buf bytes.Buffer
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(2)

		// We build a yaml.Node tree to ensure the order is openapi, info, paths, components.
		var content []*yaml.Node
		appendNode := func(key string, value interface{}) error {
			valNode := &yaml.Node{}
			if err := valNode.Encode(value); err != nil {
				return merror.Wrapf(err, "failed to encode yaml for key '%s'", key)
			}
			content = append(content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: key, Tag: "!!str"},
				valNode,
			)
			return nil
		}

		// Add fields in the desired order
		if err := appendNode("openapi", spec.OpenAPI); err != nil {
			return err
		}
		if err := appendNode("info", spec.Info); err != nil {
			return err
		}
		if spec.Paths != nil && len(spec.Paths.Map()) > 0 {
			if err := appendNode("paths", spec.Paths); err != nil {
				return err
			}
		}

		// Manually check if components are empty
		if spec.Components != nil && len(spec.Components.Schemas) > 0 {
			if err := appendNode("components", spec.Components); err != nil {
				return err
			}
		}

		root := &yaml.Node{
			Kind:    yaml.MappingNode,
			Content: content,
		}

		if err := encoder.Encode(root); err != nil {
			marshalErr = err
		} else {
			outputBytes = buf.Bytes()
		}
	}

	if marshalErr != nil {
		return merror.Wrapf(marshalErr, "failed to marshal spec to %s", format)
	}

	// Step 4: Write the output to the file.
	utils.PrintInfo("📝 Writing OpenAPI specification to {{.Path}}", utils.TplData{"Path": outputFile})
	if err := os.WriteFile(outputFile, outputBytes, 0644); err != nil {
		return merror.Wrapf(err, "failed to write OpenAPI spec to %s", outputFile)
	}

	return nil
}
