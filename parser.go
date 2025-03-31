package diffy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// HCLParser parses Terraform HCL files
type HCLParser interface {
	ParseProviderRequirements(ctx context.Context, filename string) (map[string]ProviderConfig, error)
	ParseMainFile(ctx context.Context, filename string) ([]ParsedResource, []ParsedDataSource, error)
}

// TerraformRunner runs Terraform commands
type TerraformRunner interface {
	Init(ctx context.Context, dir string) error
	GetSchema(ctx context.Context, dir string) (*TerraformSchema, error)
}

// DefaultHCLParser implements HCLParser
type DefaultHCLParser struct{}

// NewHCLParser creates a new HCL parser
func NewHCLParser() *DefaultHCLParser {
	return &DefaultHCLParser{}
}

// ParseProviderRequirements parses provider requirements from a terraform.tf file
func (p *DefaultHCLParser) ParseProviderRequirements(ctx context.Context, filename string) (map[string]ProviderConfig, error) {
	// In a real implementation, this would use HCL library to parse the file
	// For the example, we'll use a simplified approach

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return map[string]ProviderConfig{}, nil
	}

	// Mock implementation - in a real parser, you'd use hashicorp/hcl library
	providers := map[string]ProviderConfig{
		"azurerm": {
			Source:  "registry.terraform.io/hashicorp/azurerm",
			Version: "~> 3.0",
		},
		"random": {
			Source:  "registry.terraform.io/hashicorp/random",
			Version: "~> 3.1",
		},
	}

	return providers, nil
}

// ParseMainFile parses a main.tf file to extract resources and data sources
func (p *DefaultHCLParser) ParseMainFile(ctx context.Context, filename string) ([]ParsedResource, []ParsedDataSource, error) {
	// In a real implementation, this would use HCL library to parse the file
	// For the example, we'll use a simplified approach

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("file does not exist: %s", filename)
	}

	// Mock implementation - in a real parser, you'd use hashicorp/hcl library
	var resources []ParsedResource
	var dataSources []ParsedDataSource

	// Sample resource
	resources = append(resources, ParsedResource{
		Type: "azurerm_resource_group",
		Name: "example",
		Data: BlockData{
			Properties: map[string]bool{
				"name":     true,
				"location": true,
			},
			StaticBlocks:  make(map[string]*ParsedBlock),
			DynamicBlocks: make(map[string]*ParsedBlock),
		},
	})

	// Sample data source
	dataSources = append(dataSources, ParsedDataSource{
		Type: "azurerm_subscription",
		Name: "current",
		Data: BlockData{
			Properties:    make(map[string]bool),
			StaticBlocks:  make(map[string]*ParsedBlock),
			DynamicBlocks: make(map[string]*ParsedBlock),
		},
	})

	return resources, dataSources, nil
}

// NormalizeSource normalizes a provider source
func NormalizeSource(source string) string {
	if strings.Contains(source, "/") && !strings.Contains(source, "registry.terraform.io/") {
		return "registry.terraform.io/" + source
	}
	return source
}

// FindSubmodules finds submodules in a directory
func FindSubmodules(modulesDir string) ([]SubModule, error) {
	var result []SubModule

	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		return result, nil
	}

	for _, e := range entries {
		if e.IsDir() {
			subName := e.Name()
			subPath := filepath.Join(modulesDir, subName)
			mainTf := filepath.Join(subPath, "main.tf")

			if _, err := os.Stat(mainTf); err == nil {
				result = append(result, SubModule{Name: subName, Path: subPath})
			}
		}
	}

	return result, nil
}
