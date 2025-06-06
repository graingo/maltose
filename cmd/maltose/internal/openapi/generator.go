// Package openapi handles the generation of the OpenAPI specification.
package openapi

import (
	"fmt"
	"os"

	"github.com/graingo/maltose/cmd/maltose/utils"
	"gopkg.in/yaml.v3"
)

// Generate scans the given directory for Go files, parses them,
// builds an OpenAPI specification, and writes it to the output file.
func Generate(srcDir, outputFile string) error {
	utils.PrintInfo("scanningDirectory", map[string]interface{}{"Path": srcDir})

	// Step 1: Parse the source code in the directory.
	// The parser will return a structured representation of the API definitions.
	apiDefs, err := ParseDir(srcDir)
	if err != nil {
		return fmt.Errorf("failed to parse source directory: %w", err)
	}

	if len(apiDefs) == 0 {
		return fmt.Errorf("no API definitions (structs with m.Meta) found in %s", srcDir)
	}

	utils.PrintInfo("foundAPIDefinitions", map[string]interface{}{"Count": len(apiDefs)})

	// Step 2: Build the OpenAPI specification from the parsed definitions.
	spec, err := BuildSpec(apiDefs)
	if err != nil {
		return fmt.Errorf("failed to build OpenAPI spec: %w", err)
	}

	// Step 3: Marshal the specification to YAML.
	yamlData, err := yaml.Marshal(spec)
	if err != nil {
		return fmt.Errorf("failed to marshal spec to YAML: %w", err)
	}

	// Step 4: Write the YAML data to the output file.
	if err := os.WriteFile(outputFile, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write OpenAPI spec to %s: %w", outputFile, err)
	}

	return nil
}
