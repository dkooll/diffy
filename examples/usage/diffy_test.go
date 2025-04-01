package main

import (
	"testing"

	"github.com/dkooll/diffy"
)

func TestTerraformSchemaValidation(t *testing.T) {
	findings, err := diffy.ValidateSchema(
		diffy.WithTerraformRoot("../module"),
		func(opts *diffy.SchemaValidatorOptions) {
			opts.Silent = true
		},
	)

	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	for _, finding := range findings {
		t.Logf("%s", diffy.FormatFinding(finding))
	}

	if len(findings) > 0 {
		t.Errorf("Found %d missing properties/blocks in root or submodules. See logs above.", len(findings))
	}
}
