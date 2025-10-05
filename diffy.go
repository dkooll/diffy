// Package diffy validates Terraform configurations against provider schemas.
package diffy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ValidateSchema(options ...SchemaValidatorOption) ([]ValidationFinding, error) {
	opts := &SchemaValidatorOptions{
		Logger:            &SimpleLogger{},
		CreateGitHubIssue: false,
		Silent:            false,
	}

	for _, option := range options {
		option(opts)
	}

	if envRoot := os.Getenv("TERRAFORM_ROOT"); envRoot != "" {
		opts.TerraformRoot = envRoot
	}

	if envExcludedResources := os.Getenv("EXCLUDED_RESOURCES"); envExcludedResources != "" {
		resources := strings.Split(envExcludedResources, ",")
		for i, r := range resources {
			resources[i] = strings.TrimSpace(r)
		}
		opts.ExcludedResources = append(opts.ExcludedResources, resources...)
	}

	if envExcludedDataSources := os.Getenv("EXCLUDED_DATA_SOURCES"); envExcludedDataSources != "" {
		dataSources := strings.Split(envExcludedDataSources, ",")
		for i, ds := range dataSources {
			dataSources[i] = strings.TrimSpace(ds)
		}
		opts.ExcludedDataSources = append(opts.ExcludedDataSources, dataSources...)
	}

	if opts.TerraformRoot == "" {
		return nil, fmt.Errorf("terraform root path not specified - set TERRAFORM_ROOT environment variable or use WithTerraformRoot option")
	}

	findings, err := validateProject(opts)
	if err != nil {
		return nil, err
	}

	if !opts.Silent {
		outputFindings(findings)
	}

	if opts.CreateGitHubIssue {
		ctx := context.Background()
		if err := createGitHubIssue(ctx, opts, findings); err != nil {
			opts.Logger.Logf("Failed to create/update GitHub issue: %v", err)
		}
	}

	return findings, nil
}

func validateProject(opts *SchemaValidatorOptions) ([]ValidationFinding, error) {
	absRoot, err := filepath.Abs(opts.TerraformRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for %s: %w", opts.TerraformRoot, err)
	}

	rootFindings, err := ValidateTerraformSchemaInDirectoryWithOptions(opts.Logger, absRoot, "", opts.ExcludedResources, opts.ExcludedDataSources)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	var allFindings []ValidationFinding
	allFindings = append(allFindings, rootFindings...)

	modulesDir := filepath.Join(absRoot, "modules")
	submodules, err := FindSubmodules(modulesDir)
	if err != nil {
		if !opts.Silent {
			fmt.Printf("Note: No submodules found in %s\n", modulesDir)
		}
	} else {
		for _, sm := range submodules {
			findings, err := ValidateTerraformSchemaInDirectoryWithOptions(opts.Logger, sm.Path, sm.Name, opts.ExcludedResources, opts.ExcludedDataSources)
			if err != nil {
				opts.Logger.Logf("Failed to validate submodule %s: %v", sm.Name, err)
				continue
			}
			allFindings = append(allFindings, findings...)
		}
	}

	deduplicatedFindings := DeduplicateFindings(allFindings)

	return deduplicatedFindings, nil
}

func outputFindings(findings []ValidationFinding) {
	if len(findings) == 0 {
		fmt.Println("No validation findings.")
		return
	}

	fmt.Printf("Found %d issues:\n", len(findings))

	for _, finding := range findings {
		fmt.Println(FormatFinding(finding))
	}
}

func createGitHubIssue(ctx context.Context, opts *SchemaValidatorOptions, findings []ValidationFinding) error {
	if opts.GitHubToken == "" {
		return fmt.Errorf("GitHub token not provided")
	}

	owner := opts.GitHubOwner
	repo := opts.GitHubRepo

	if owner == "" || repo == "" {
		gi := NewGitRepoInfo(opts.TerraformRoot)
		owner, repo = gi.GetRepoInfo()
		if owner == "" || repo == "" {
			return fmt.Errorf("could not determine repository info for GitHub issue creation")
		}
	}

	issueManager := NewGitHubIssueManager(owner, repo, opts.GitHubToken)

	if len(findings) == 0 {
		return issueManager.CloseExistingIssuesIfEmpty(ctx)
	}

	return issueManager.CreateOrUpdateIssue(ctx, findings)
}
